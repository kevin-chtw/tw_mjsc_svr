#!/usr/bin/env python3
"""
麻将AI训练服务 - 使用HTTP REST API
支持DQN训练（可选PyTorch，无GPU也能运行）
"""

from http.server import HTTPServer, BaseHTTPRequestHandler
import json
import numpy as np
import random
from collections import deque
from datetime import datetime
import sys
import logging

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s [%(levelname)s] %(message)s',
    handlers=[
        logging.FileHandler('ai_service.log'),
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger(__name__)

try:
    import torch
    import torch.nn as nn
    import torch.optim as optim
    HAS_TORCH = True
    logger.info("✅ PyTorch available")
except ImportError:
    HAS_TORCH = False
    logger.warning("⚠️  PyTorch not available, using random policy")

class DQN(nn.Module if HAS_TORCH else object):
    """Dueling DQN 网络 - 适合麻将AI
    
    特点：
    1. Dueling架构：分离状态价值V(s)和动作优势A(s,a)
    2. 残差连接：提高深层网络的训练效果
    3. LayerNorm：稳定训练过程
    4. 合理的网络深度：足够表达复杂策略，但不会过拟合
    """
    def __init__(self, input_dim=3185, hidden_dim=512, output_dim=137):
        if HAS_TORCH:
            super().__init__()
            
            # 共享特征提取层（残差块）
            self.shared = nn.Sequential(
                nn.Linear(input_dim, hidden_dim),
                nn.LayerNorm(hidden_dim),
                nn.ReLU(),
                nn.Dropout(0.1),
            )
            
            # 残差块1
            self.residual1 = nn.Sequential(
                nn.Linear(hidden_dim, hidden_dim),
                nn.LayerNorm(hidden_dim),
                nn.ReLU(),
                nn.Dropout(0.1),
                nn.Linear(hidden_dim, hidden_dim),
                nn.LayerNorm(hidden_dim),
            )
            
            # 残差块2
            self.residual2 = nn.Sequential(
                nn.Linear(hidden_dim, hidden_dim),
                nn.LayerNorm(hidden_dim),
                nn.ReLU(),
                nn.Dropout(0.1),
                nn.Linear(hidden_dim, hidden_dim),
                nn.LayerNorm(hidden_dim),
            )
            
            # Dueling架构：状态价值流
            self.value_stream = nn.Sequential(
                nn.Linear(hidden_dim, hidden_dim // 2),
                nn.ReLU(),
                nn.Linear(hidden_dim // 2, 1)
            )
            
            # Dueling架构：动作优势流
            self.advantage_stream = nn.Sequential(
                nn.Linear(hidden_dim, hidden_dim // 2),
                nn.ReLU(),
                nn.Linear(hidden_dim // 2, output_dim)
            )
        
    def forward(self, x):
        if not HAS_TORCH:
            return None
            
        # 共享特征提取
        shared_features = self.shared(x)
        
        # 残差连接1
        residual = self.residual1(shared_features)
        shared_features = torch.relu(shared_features + residual)
        
        # 残差连接2
        residual = self.residual2(shared_features)
        shared_features = torch.relu(shared_features + residual)
        
        # Dueling架构：计算V(s)和A(s,a)
        value = self.value_stream(shared_features)
        advantages = self.advantage_stream(shared_features)
        
        # Q(s,a) = V(s) + (A(s,a) - mean(A(s,a)))
        # 减去平均值是为了唯一性（identifiability）
        q_values = value + (advantages - advantages.mean(dim=1, keepdim=True))
        
        return q_values

class AIService:
    def __init__(self):
        self.device = "cpu"
        if HAS_TORCH:
            self.model = DQN()
            self.target_model = DQN()
            self.target_model.load_state_dict(self.model.state_dict())
            # 使用AdamW优化器（带权重衰减，防止过拟合）
            # 降低权重衰减，避免过度正则化
            self.optimizer = optim.AdamW(self.model.parameters(), lr=0.0003, weight_decay=0.001)
            # 使用Huber Loss（对异常值更鲁棒）
            self.criterion = nn.SmoothL1Loss()
            # 学习率调度器（更慢的衰减，保持学习能力）
            # 每2000步衰减一次，而不是1000步
            self.scheduler = optim.lr_scheduler.StepLR(self.optimizer, step_size=2000, gamma=0.95)
        else:
            self.model = None
            self.target_model = None
        
        # 更大的replay buffer，存储更多经验
        self.replay_buffer = deque(maxlen=50000)
        # epsilon参数
        self.epsilon = 1.0  # 从高探索开始
        self.epsilon_min = 0.15  # 提高最小值，保持更多探索（从0.1提升到0.15）
        self.epsilon_decay = 0.9995  # 更缓慢的衰减
        # 训练参数
        self.gamma = 0.99
        self.batch_size = 128  # 更大的batch size，训练更稳定
        self.train_count = 0
        self.update_target_every = 100  # 更频繁更新目标网络
        self.save_every = 1000  # 每1000步自动保存一次模型
        self.last_save_count = 0  # 上次保存的训练步数
        # 优先经验回放参数（可选）
        self.use_prioritized_replay = False  # 暂时关闭，简化实现
        
        logger.info(f"🚀 AI Service initialized (PyTorch: {HAS_TORCH})")
        logger.info(f"   Model params: {sum(p.numel() for p in self.model.parameters()) if HAS_TORCH else 0:,}")
    
    def get_decision(self, obs, candidates):
        """从候选动作中选择最佳动作"""
        if not candidates:
            return {'operate': 1, 'tile': 0}  # OPERATE_PASS
        
        # 检查PASS候选的tile值
        for cand in candidates:
            if cand['operate'] == 1 and cand['tile'] != 0:
                logger.warning(f"⚠️  Received PASS candidate with tile={cand['tile']}, should be 0!")
        
        # 使用DQN选择最佳动作
        if HAS_TORCH and random.random() > self.epsilon:
            # 计算所有候选动作的Q值
            obs_tensor = torch.FloatTensor(obs).unsqueeze(0)
            with torch.no_grad():
                q_values = self.model(obs_tensor).squeeze(0).numpy()
            
            best_candidate = None
            best_q = float('-inf')
            
            for cand in candidates:
                action_idx = self._get_action_index(cand['operate'], cand['tile'])
                if action_idx is not None and action_idx < len(q_values):
                    q = q_values[action_idx]
                    if q > best_q:
                        best_q = q
                        best_candidate = cand
            
            if best_candidate:
                return self._normalize_decision(best_candidate)
        
        # 随机选择或DQN失败时的fallback
        selected = random.choice(candidates)
        return self._normalize_decision(selected)
    
    def _normalize_decision(self, decision):
        """标准化决策：PASS操作的tile统一为0"""
        OPERATE_PASS = 1
        original_tile = decision['tile']
        if decision['operate'] == OPERATE_PASS:
            if original_tile != 0:
                logger.warning(f"⚠️  Normalizing PASS decision: tile {original_tile} -> 0")
            return {'operate': OPERATE_PASS, 'tile': 0}
        return decision    
  
    def _get_action_index(self, operate, tile):
        """计算动作索引（0-136）
        注意：Go侧已经将tile转换为index(0-33)再传过来
        """
        OPERATE_PASS = 1
        OPERATE_PON = 4
        OPERATE_KON = 8
        OPERATE_HU = 32
        OPERATE_DISCARD = 64
        
        # tile已经是index(0-33)，直接使用
        if operate == OPERATE_DISCARD:
            return tile
        elif operate == OPERATE_PON:
            return 34 + tile
        elif operate == OPERATE_KON:
            return 68 + tile
        elif operate == OPERATE_HU:
            return 102 + tile
        elif operate == OPERATE_PASS:
            return 136
        return None
    
    def get_action(self, state, valid_actions):
        """推理：根据状态和有效动作返回最佳动作（旧接口）"""
        if not valid_actions:
            return 0
        
        # ε-greedy
        if random.random() < self.epsilon or not HAS_TORCH:
            return random.choice(valid_actions)
        
        with torch.no_grad():
            state_tensor = torch.FloatTensor(state).unsqueeze(0)
            q_values = self.model(state_tensor).squeeze(0).numpy()
            
            # 只考虑有效动作
            valid_q = [(idx, q_values[idx]) for idx in valid_actions]
            return max(valid_q, key=lambda x: x[1])[0]
    
    def report_episode(self, episode_data):
        """训练：接收一局轨迹"""
        steps = episode_data.get('steps', [])
        
        # 将轨迹加入 replay buffer
        for step in steps:
            state = np.array(step['state'], dtype=np.float32)
            next_state = np.array(step.get('next_state', []), dtype=np.float32) if step.get('next_state') else None
            
            # 从operate和tile计算action_idx
            action_idx = self._get_action_index(step['operate'], step['tile'])
            if action_idx is None:
                continue  # 跳过无效动作
            
            self.replay_buffer.append({
                'state': state,
                'action': action_idx,
                'reward': step['reward'],
                'next_state': next_state,
                'done': step.get('done', False)
            })
        
        # 如果buffer足够大且有PyTorch，进行训练
        if HAS_TORCH and len(self.replay_buffer) >= self.batch_size:
            loss = self._train()
            
            # 定期更新目标网络
            if self.train_count % self.update_target_every == 0:
                self.target_model.load_state_dict(self.model.state_dict())
                logger.info(f"🔄 Updated target network at train step {self.train_count}")
        
        is_hu = episode_data.get('is_hu', False)
        hu_multi = episode_data.get('hu_multi', 0)
        shaped_reward = episode_data.get('shaped_reward', 0)
        
        logger.info(f"📦 Episode: {len(steps)} steps, buffer: {len(self.replay_buffer)}, hu: {is_hu}, multi: {hu_multi}, reward: {shaped_reward:.2f}")
        return {'status': 'ok'}
    
    def _train(self):
        """从 replay buffer 采样并训练"""
        if not HAS_TORCH or len(self.replay_buffer) < self.batch_size:
            return 0.0
        
        batch = random.sample(self.replay_buffer, self.batch_size)
        
        # 使用 numpy.stack 提高转换效率
        states_np = np.stack([t['state'] for t in batch])
        actions_np = np.array([t['action'] for t in batch], dtype=np.int64)
        rewards_np = np.array([t['reward'] for t in batch], dtype=np.float32)
        next_states_np = np.stack([
            t['next_state'] if t['next_state'] is not None else np.zeros_like(t['state']) 
            for t in batch
        ])
        dones_np = np.array([t['done'] for t in batch], dtype=np.float32)
        
        # 转换为 tensor（更快）
        states = torch.from_numpy(states_np)
        actions = torch.from_numpy(actions_np)
        rewards = torch.from_numpy(rewards_np)
        next_states = torch.from_numpy(next_states_np)
        dones = torch.from_numpy(dones_np)
        
        # 计算当前 Q 值
        current_q = self.model(states).gather(1, actions.unsqueeze(1)).squeeze(1)
        
        # 正确的 Double DQN 实现：
        # 1. 先用主网络选择动作
        # 2. 再用目标网络评估Q值
        # 这样可以避免Q值高估
        with torch.no_grad():
            next_actions = self.model(next_states).max(1)[1].unsqueeze(1)
            next_q = self.target_model(next_states).gather(1, next_actions).squeeze(1)
            target_q = rewards + self.gamma * next_q * (1 - dones)
        
        # 计算损失并更新
        loss = self.criterion(current_q, target_q)
        self.optimizer.zero_grad()
        loss.backward()
        
        # 梯度裁剪（防止梯度爆炸）
        # 降低max_norm，使梯度更新更敏感
        grad_norm = torch.nn.utils.clip_grad_norm_(self.model.parameters(), max_norm=1.0)
        
        self.optimizer.step()
        
        # 监控梯度（每100次训练记录一次）
        if self.train_count % 100 == 0 and grad_norm < 0.001:
            logger.warning(f"⚠️  Gradient too small: {grad_norm:.6f}, model may not be learning!")
        
        # epsilon 衰减（指数衰减到最小值）
        self.epsilon = max(self.epsilon_min, self.epsilon * self.epsilon_decay)
        
        # 学习率调度
        self.scheduler.step()
        
        self.train_count += 1
        
        # 定期打印训练信息
        if self.train_count % 20 == 0:
            current_lr = self.optimizer.param_groups[0]['lr']
            # 计算平均Q值，监控模型输出
            avg_q = current_q.mean().item()
            avg_target_q = target_q.mean().item()
            logger.info(f"🔥 Train #{self.train_count}: loss={loss.item():.6f}, epsilon={self.epsilon:.3f}, lr={current_lr:.6f}, buffer={len(self.replay_buffer)}, avg_q={avg_q:.3f}, avg_target_q={avg_target_q:.3f}")
        
        # 自动保存模型
        if self.train_count - self.last_save_count >= self.save_every:
            self.save_model()
            # 同时保存一个带时间戳的备份
            backup_path = f'mahjong_dqn_backup_{datetime.now().strftime("%Y%m%d_%H%M%S")}.pth'
            self.save_model(backup_path)
            self.last_save_count = self.train_count
            logger.info(f"💾 Auto-saved model (train_count={self.train_count})")
        
        return loss.item()
    
    def save_model(self, path='mahjong_dqn.pth'):
        """保存模型"""
        if HAS_TORCH:
            current_lr = self.optimizer.param_groups[0]['lr']
            torch.save({
                'model_state_dict': self.model.state_dict(),
                'target_model_state_dict': self.target_model.state_dict(),
                'optimizer_state_dict': self.optimizer.state_dict(),
                'scheduler_state_dict': self.scheduler.state_dict(),
                'epsilon': self.epsilon,
                'train_count': self.train_count,
                'buffer_size': len(self.replay_buffer),
                'learning_rate': current_lr,  # 保存当前学习率
            }, path)
            logger.info(f"💾 Model saved to {path} (train_count={self.train_count}, epsilon={self.epsilon:.3f}, lr={current_lr:.8f})")
    
    def load_model(self, path='mahjong_dqn.pth', reset_lr=True, reset_epsilon=True):
        """加载模型
        
        Args:
            path: 模型文件路径
            reset_lr: 是否重置学习率（默认True，重置到0.0002）
            reset_epsilon: 是否重置探索率（默认True，重置到0.2）
        """
        if HAS_TORCH:
            try:
                checkpoint = torch.load(path, weights_only=True)
                self.model.load_state_dict(checkpoint['model_state_dict'])
                self.target_model.load_state_dict(checkpoint['target_model_state_dict'])
                self.optimizer.load_state_dict(checkpoint['optimizer_state_dict'])
                
                # 重置学习率（如果学习率过低）
                if reset_lr:
                    current_lr = self.optimizer.param_groups[0]['lr']
                    if current_lr < 0.0001:  # 如果学习率过低，重置
                        new_lr = 0.0002  # 重置到0.0002
                        for param_group in self.optimizer.param_groups:
                            param_group['lr'] = new_lr
                        logger.info(f"🔄 Learning rate reset: {current_lr:.8f} -> {new_lr:.6f}")
                    else:
                        logger.info(f"✅ Learning rate OK: {current_lr:.6f}")
                
                # 重置学习率调度器（使用新的学习率）
                if reset_lr:
                    self.scheduler = optim.lr_scheduler.StepLR(self.optimizer, step_size=2000, gamma=0.95)
                    logger.info(f"🔄 Learning rate scheduler reset")
                elif 'scheduler_state_dict' in checkpoint:
                    self.scheduler.load_state_dict(checkpoint['scheduler_state_dict'])
                
                # 重置探索率
                if reset_epsilon:
                    self.epsilon = 0.2  # 重置到0.2，保持一定探索
                    logger.info(f"🔄 Epsilon reset to: {self.epsilon:.3f}")
                else:
                    self.epsilon = checkpoint.get('epsilon', 0.2)
                    # 确保epsilon不低于最小值
                    self.epsilon = max(self.epsilon_min, self.epsilon)
                
                self.train_count = checkpoint.get('train_count', 0)
                self.last_save_count = self.train_count  # 恢复上次保存的步数
                buffer_size = checkpoint.get('buffer_size', 0)
                logger.info(f"📂 Model loaded from {path}")
                logger.info(f"   Train count: {self.train_count}, Epsilon: {self.epsilon:.3f}, Buffer was: {buffer_size}")
            except FileNotFoundError:
                logger.warning(f"⚠️  Model file not found: {path}, starting fresh")
            except Exception as e:
                logger.warning(f"⚠️  Error loading model: {e}, starting fresh")

# 全局服务实例
ai_service = AIService()
# 尝试加载已有模型
ai_service.load_model()

class RequestHandler(BaseHTTPRequestHandler):
    def log_message(self, format, *args):
        # 减少日志输出
        pass
    
    def do_POST(self):
        content_length = int(self.headers.get('Content-Length', 0))
        post_data = self.rfile.read(content_length) if content_length > 0 else b'{}'
        
        try:
            data = json.loads(post_data.decode('utf-8')) if post_data else {}
            
            if self.path == '/get_action':
                # GetAction 接口
                state = data['state']
                valid_actions = data['valid_actions']
                action_idx = ai_service.get_action(state, valid_actions)
                
                response = {'action_idx': int(action_idx)}
                self.send_response(200)
                self.send_header('Content-Type', 'application/json')
                self.end_headers()
                self.wfile.write(json.dumps(response).encode())
                
            elif self.path == '/get_decision':
                # GetDecision 接口 - 从候选动作中选择最佳动作
                obs = data['obs']
                candidates = data['candidates']  # [{operate, tile}, ...]
                
                decision = ai_service.get_decision(obs, candidates)
                
                self.send_response(200)
                self.send_header('Content-Type', 'application/json')
                self.end_headers()
                self.wfile.write(json.dumps(decision).encode())
            
            elif self.path == '/report_episode':
                # ReportEpisode 接口
                result = ai_service.report_episode(data)
                
                self.send_response(200)
                self.send_header('Content-Type', 'application/json')
                self.end_headers()
                self.wfile.write(json.dumps(result).encode())
                
            elif self.path == '/save_model':
                # 保存模型接口
                ai_service.save_model()
                
                self.send_response(200)
                self.send_header('Content-Type', 'application/json')
                self.end_headers()
                self.wfile.write(json.dumps({'status': 'saved'}).encode())
            else:
                self.send_error(404)
        
        except Exception as e:
            logger.error(f"❌ Error: {e}")
            import traceback
            traceback.print_exc()
            self.send_error(500, str(e))
    
    def do_GET(self):
        if self.path == '/health':
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps({
                'status': 'healthy',
                'pytorch': HAS_TORCH,
                'buffer_size': len(ai_service.replay_buffer),
                'epsilon': ai_service.epsilon,
                'train_count': ai_service.train_count
            }).encode())
        else:
            self.send_error(404)

def serve(port=50051):
    server = HTTPServer(('0.0.0.0', port), RequestHandler)
    logger.info(f"✅ AI Service listening on http://0.0.0.0:{port}")
    logger.info(f"   GET  /health - 健康检查")
    logger.info(f"   POST /get_decision - 获取决策")
    logger.info(f"   POST /report_episode - 上报轨迹")
    logger.info(f"   POST /save_model - 保存模型")
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        logger.info("\n🛑 Shutting down...")
        # 退出前保存模型
        ai_service.save_model()
        server.shutdown()

if __name__ == '__main__':
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 50051
    serve(port)

