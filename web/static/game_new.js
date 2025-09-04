class LiveOpsPlatform {
    constructor() {
        // 기본 설정
        this.playerId = null;
        this.playerData = {
            level: 1,
            xp: 0,
            totalScore: 0,
            achievements: [],
            stats: {}
        };
        
        // 게임 데이터
        this.games = {
            clicker: {
                score: 0,
                clickCount: 0,
                multiplier: 1,
                upgradeCost: 100
            },
            memory: {
                level: 1,
                score: 0,
                sequence: [],
                currentSequence: []
            },
            reaction: {
                bestTime: null,
                times: [],
                isWaiting: false,
                startTime: null
            }
        };
        
        // 서비스 URL 설정
        this.services = {
            gateway: 'http://localhost:8080',
            leaderboard: 'http://localhost:8082',
            progression: 'http://localhost:8083',
            fairness: 'http://localhost:8081'
        };
        
        // 현재 탭
        this.currentTab = 'dashboard';
        this.currentGame = null;
        
        this.init();
    }
    
    async init() {
        await this.initializePlayer();
        this.setupEventListeners();
        this.startPeriodicUpdates();
        this.checkServiceHealth();
        this.showNotification('플랫폼이 초기화되었습니다.', 'success');
    }
    
    // 플레이어 초기화
    async initializePlayer() {
        try {
            // IP 기반 플레이어 ID 생성
            this.playerId = 'player_' + this.generateRandomId();
            document.getElementById('playerId').textContent = this.playerId;
            
            // 플레이어 진행도 로드
            await this.loadPlayerProgress();
            await this.updateDashboard();
            
        } catch (error) {
            console.error('플레이어 초기화 실패:', error);
            this.playerId = 'offline_' + Date.now();
            document.getElementById('playerId').textContent = this.playerId + ' (오프라인)';
        }
    }
    
    generateRandomId() {
        return Math.random().toString(36).substring(2, 15);
    }
    
    // 이벤트 리스너 설정
    setupEventListeners() {
        // 탭 전환
        document.querySelectorAll('.nav-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                this.switchTab(e.target.dataset.tab);
            });
        });
        
        // 게임 카드 클릭
        document.querySelectorAll('.game-card').forEach(card => {
            card.addEventListener('click', (e) => {
                this.startGame(e.currentTarget.dataset.game);
            });
        });
        
        // 게임으로 돌아가기
        document.getElementById('backToGames').addEventListener('click', () => {
            this.backToGameSelection();
        });
        
        // 클리커 게임
        document.getElementById('mainClickBtn').addEventListener('click', () => {
            this.handleClick();
        });
        
        document.getElementById('buyMultiplier').addEventListener('click', () => {
            this.buyUpgrade();
        });
        
        // 메모리 게임
        document.getElementById('startMemory').addEventListener('click', () => {
            this.startMemoryGame();
        });
        
        document.getElementById('submitMemory').addEventListener('click', () => {
            this.submitMemoryAnswer();
        });
        
        // 반응속도 게임
        document.getElementById('startReaction').addEventListener('click', () => {
            this.startReactionGame();
        });
        
        // 리더보드 필터
        document.querySelectorAll('.filter-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                this.filterLeaderboard(e.target.dataset.type);
            });
        });
        
        // 보상 받기
        document.getElementById('claimAllRewards').addEventListener('click', () => {
            this.claimAllRewards();
        });
    }
    
    // 탭 전환
    switchTab(tabName) {
        // 네비게이션 업데이트
        document.querySelectorAll('.nav-btn').forEach(btn => {
            btn.classList.remove('active');
        });
        document.querySelector(`[data-tab="${tabName}"]`).classList.add('active');
        
        // 컨텐츠 업데이트
        document.querySelectorAll('.tab-content').forEach(content => {
            content.classList.remove('active');
        });
        document.getElementById(tabName).classList.add('active');
        
        this.currentTab = tabName;
        
        // 탭별 데이터 로드
        switch (tabName) {
            case 'dashboard':
                this.updateDashboard();
                break;
            case 'leaderboard':
                this.loadLeaderboard();
                break;
            case 'progression':
                this.updateProgression();
                break;
            case 'admin':
                this.updateMonitoring();
                break;
        }
    }
    
    // 게임 시작
    startGame(gameType) {
        this.currentGame = gameType;
        document.getElementById('gamePlayArea').classList.remove('hidden');
        document.querySelector('.games-grid').style.display = 'none';
        
        // 게임별 초기화
        switch (gameType) {
            case 'clicker':
                document.getElementById('currentGameTitle').textContent = '클리커 게임';
                document.getElementById('clickerGame').classList.remove('hidden');
                this.updateClickerDisplay();
                break;
            case 'memory':
                document.getElementById('currentGameTitle').textContent = '메모리 게임';
                document.getElementById('memoryGame').classList.remove('hidden');
                this.updateMemoryDisplay();
                break;
            case 'reaction':
                document.getElementById('currentGameTitle').textContent = '반응속도 게임';
                document.getElementById('reactionGame').classList.remove('hidden');
                this.updateReactionDisplay();
                break;
        }
    }
    
    // 게임 선택으로 돌아가기
    backToGameSelection() {
        document.getElementById('gamePlayArea').classList.add('hidden');
        document.querySelector('.games-grid').style.display = 'grid';
        
        // 모든 게임 컨텐츠 숨기기
        document.querySelectorAll('.game-content').forEach(content => {
            content.classList.add('hidden');
        });
        
        this.currentGame = null;
    }
    
    // 클리커 게임 로직
    async handleClick() {
        const clickValue = 10 * this.games.clicker.multiplier;
        this.games.clicker.score += clickValue;
        this.games.clicker.clickCount++;
        this.playerData.totalScore += clickValue;
        
        // XP 획득
        const xpGained = 10;
        await this.gainXP(xpGained);
        
        // 이벤트 전송
        await this.sendGameEvent('clicker_click', {
            deltaScore: clickValue,
            totalClicks: this.games.clicker.clickCount,
            multiplier: this.games.clicker.multiplier
        });
        
        this.updateClickerDisplay();
        this.showXPGain(xpGained);
    }
    
    buyUpgrade() {
        if (this.games.clicker.score >= this.games.clicker.upgradeCost) {
            this.games.clicker.score -= this.games.clicker.upgradeCost;
            this.games.clicker.multiplier++;
            this.games.clicker.upgradeCost = Math.floor(this.games.clicker.upgradeCost * 1.5);
            
            this.updateClickerDisplay();
            this.showNotification(`배율이 ${this.games.clicker.multiplier}배로 증가했습니다!`, 'success');
        }
    }
    
    updateClickerDisplay() {
        document.getElementById('clickerScore').textContent = this.games.clicker.score;
        document.getElementById('clickerMultiplier').textContent = this.games.clicker.multiplier;
        document.getElementById('clickCount').textContent = this.games.clicker.clickCount;
        
        const upgradeBtn = document.getElementById('buyMultiplier');
        upgradeBtn.textContent = `배율 증가 (${this.games.clicker.upgradeCost}점)`;
        upgradeBtn.disabled = this.games.clicker.score < this.games.clicker.upgradeCost;
    }
    
    // 메모리 게임 로직
    startMemoryGame() {
        this.games.memory.sequence = this.generateMemorySequence(this.games.memory.level + 2);
        this.displayMemorySequence();
        document.getElementById('startMemory').disabled = true;
        document.getElementById('memoryInput').value = '';
    }
    
    generateMemorySequence(length) {
        const sequence = [];
        for (let i = 0; i < length; i++) {
            sequence.push(Math.floor(Math.random() * 9) + 1);
        }
        return sequence;
    }
    
    async displayMemorySequence() {
        const sequenceEl = document.getElementById('memorySequence');
        sequenceEl.textContent = '준비하세요...';
        
        await this.sleep(1000);
        
        for (let i = 0; i < this.games.memory.sequence.length; i++) {
            sequenceEl.textContent = this.games.memory.sequence.slice(0, i + 1).join(' ');
            await this.sleep(800);
        }
        
        await this.sleep(2000);
        sequenceEl.textContent = '?';
        document.getElementById('submitMemory').disabled = false;
    }
    
    async submitMemoryAnswer() {
        const input = document.getElementById('memoryInput').value.replace(/\s/g, '');
        const correct = this.games.memory.sequence.join('');
        
        if (input === correct) {
            const score = this.games.memory.level * 50;
            const xpGained = 50;
            
            this.games.memory.score += score;
            this.games.memory.level++;
            this.playerData.totalScore += score;
            
            await this.gainXP(xpGained);
            await this.sendGameEvent('memory_success', {
                level: this.games.memory.level - 1,
                deltaScore: score,
                sequence: this.games.memory.sequence
            });
            
            this.showNotification(`정답! 레벨 ${this.games.memory.level}`, 'success');
            this.showXPGain(xpGained);
        } else {
            this.showNotification('틀렸습니다. 다시 시도하세요.', 'error');
            this.games.memory.level = Math.max(1, this.games.memory.level - 1);
        }
        
        this.updateMemoryDisplay();
        document.getElementById('startMemory').disabled = false;
        document.getElementById('submitMemory').disabled = true;
    }
    
    updateMemoryDisplay() {
        document.getElementById('memoryLevel').textContent = this.games.memory.level;
        document.getElementById('memoryScore').textContent = this.games.memory.score;
    }
    
    // 반응속도 게임 로직
    async startReactionGame() {
        document.getElementById('startReaction').disabled = true;
        document.getElementById('reactionMessage').textContent = '잠시 기다리세요...';
        document.getElementById('reactionTarget').classList.add('hidden');
        
        // 랜덤 대기 시간 (2-5초)
        const waitTime = Math.random() * 3000 + 2000;
        await this.sleep(waitTime);
        
        document.getElementById('reactionMessage').textContent = '';
        document.getElementById('reactionTarget').classList.remove('hidden');
        this.games.reaction.startTime = Date.now();
        this.games.reaction.isWaiting = true;
        
        document.getElementById('reactionTarget').addEventListener('click', this.handleReactionClick.bind(this), { once: true });
    }
    
    async handleReactionClick() {
        if (!this.games.reaction.isWaiting) return;
        
        const reactionTime = Date.now() - this.games.reaction.startTime;
        this.games.reaction.times.push(reactionTime);
        this.games.reaction.isWaiting = false;
        
        if (!this.games.reaction.bestTime || reactionTime < this.games.reaction.bestTime) {
            this.games.reaction.bestTime = reactionTime;
            this.showNotification('새 기록!', 'success');
        }
        
        const xpGained = Math.max(10, Math.floor(50 - (reactionTime / 20)));
        await this.gainXP(xpGained);
        await this.sendGameEvent('reaction_success', {
            reactionTime: reactionTime,
            bestTime: this.games.reaction.bestTime
        });
        
        document.getElementById('reactionTarget').classList.add('hidden');
        document.getElementById('reactionMessage').textContent = `반응 시간: ${reactionTime}ms`;
        document.getElementById('startReaction').disabled = false;
        
        this.updateReactionDisplay();
        this.showXPGain(xpGained);
    }
    
    updateReactionDisplay() {
        document.getElementById('bestTime').textContent = this.games.reaction.bestTime || '-';
        const avgTime = this.games.reaction.times.length > 0 
            ? Math.floor(this.games.reaction.times.reduce((a, b) => a + b) / this.games.reaction.times.length)
            : '-';
        document.getElementById('avgTime').textContent = avgTime;
    }
    
    // XP 획득
    async gainXP(amount) {
        this.playerData.xp += amount;
        
        // 레벨업 체크
        const levelThreshold = this.playerData.level * 100 + (this.playerData.level - 1) * 50;
        if (this.playerData.xp >= levelThreshold) {
            await this.levelUp();
        }
        
        await this.sendProgressionEvent('xp_gain', { deltaXp: amount });
        this.updateXPDisplay();
    }
    
    async levelUp() {
        this.playerData.level++;
        this.playerData.xp = 0; // 간단한 구현
        
        this.showNotification(`레벨 업! 현재 레벨: ${this.playerData.level}`, 'success');
        await this.checkAchievements();
    }
    
    // 업적 체크
    async checkAchievements() {
        const newAchievements = [];
        
        if (this.playerData.level >= 5 && !this.playerData.achievements.includes('level_5')) {
            newAchievements.push('level_5');
            this.playerData.achievements.push('level_5');
        }
        
        if (this.games.clicker.clickCount >= 100 && !this.playerData.achievements.includes('clicker_100')) {
            newAchievements.push('clicker_100');
            this.playerData.achievements.push('clicker_100');
        }
        
        newAchievements.forEach(achievement => {
            this.showAchievement(achievement);
        });
    }
    
    // 대시보드 업데이트
    async updateDashboard() {
        document.getElementById('playerLevel').textContent = this.playerData.level;
        document.getElementById('playerXP').textContent = this.playerData.xp;
        document.getElementById('totalScore').textContent = this.playerData.totalScore;
        document.getElementById('achievementsCount').textContent = this.playerData.achievements.length;
        
        // 게임 통계
        const totalGames = this.games.clicker.clickCount + this.games.memory.level + this.games.reaction.times.length;
        document.getElementById('gamesPlayed').textContent = totalGames;
        document.getElementById('totalXPGained').textContent = this.playerData.xp + (this.playerData.level - 1) * 100;
        
        this.updateXPDisplay();
        this.updateAchievementsList();
        
        // 랭크 조회
        try {
            const rank = await this.getPlayerRank();
            document.getElementById('currentRank').textContent = rank || '-';
        } catch (error) {
            document.getElementById('currentRank').textContent = '-';
        }
    }
    
    updateXPDisplay() {
        const levelThreshold = this.playerData.level * 100 + (this.playerData.level - 1) * 50;
        const progress = (this.playerData.xp / levelThreshold) * 100;
        
        document.getElementById('xpProgress').style.width = progress + '%';
        document.getElementById('xpText').textContent = `${this.playerData.xp} / ${levelThreshold} XP`;
        
        // 진행도 탭의 XP 바도 업데이트
        document.getElementById('progLevel').textContent = this.playerData.level;
        document.getElementById('progXPBar').style.width = progress + '%';
        document.getElementById('progXPText').textContent = `${this.playerData.xp} / ${levelThreshold} XP`;
    }
    
    updateAchievementsList() {
        const list = document.getElementById('achievementsList');
        const achievements = {
            'level_5': { icon: 'fas fa-star', name: '레벨 5 달성' },
            'clicker_100': { icon: 'fas fa-mouse', name: '100번 클릭' }
        };
        
        if (this.playerData.achievements.length === 0) {
            list.innerHTML = '<div class="no-achievements">아직 획득한 업적이 없습니다</div>';
        } else {
            list.innerHTML = this.playerData.achievements.map(id => {
                const achievement = achievements[id];
                return `
                    <div class="achievement-item">
                        <i class="${achievement.icon}"></i>
                        ${achievement.name}
                    </div>
                `;
            }).join('');
        }
    }
    
    // 리더보드 로드
    async loadLeaderboard(type = 'global') {
        try {
            this.showLoading('leaderboardBody');
            
            const response = await fetch(`${this.services.leaderboard}/board?limit=10`);
            const data = await response.json();
            
            const tbody = document.getElementById('leaderboardBody');
            if (data.entries && data.entries.length > 0) {
                tbody.innerHTML = data.entries.map((entry, index) => {
                    const isCurrentPlayer = entry.playerId === this.playerId;
                    const rankClass = index === 0 ? 'gold' : index === 1 ? 'silver' : index === 2 ? 'bronze' : '';
                    
                    return `
                        <div class="leaderboard-row ${isCurrentPlayer ? 'current-player' : ''}">
                            <div class="rank ${rankClass}">#${entry.rank || index + 1}</div>
                            <div>${entry.playerName || entry.playerId}</div>
                            <div>${entry.score.toLocaleString()}</div>
                            <div>${entry.level || '-'}</div>
                            <div>${new Date(entry.updatedAt).toLocaleString()}</div>
                        </div>
                    `;
                }).join('');
            } else {
                tbody.innerHTML = '<div class="loading">리더보드가 비어있습니다</div>';
            }
        } catch (error) {
            console.error('리더보드 로드 실패:', error);
            document.getElementById('leaderboardBody').innerHTML = '<div class="loading">리더보드를 불러올 수 없습니다</div>';
        }
    }
    
    async getPlayerRank() {
        try {
            const response = await fetch(`${this.services.leaderboard}/player/${this.playerId}/rank`);
            const data = await response.json();
            return data.rank;
        } catch (error) {
            return null;
        }
    }
    
    filterLeaderboard(type) {
        document.querySelectorAll('.filter-btn').forEach(btn => {
            btn.classList.remove('active');
        });
        document.querySelector(`[data-type="${type}"]`).classList.add('active');
        this.loadLeaderboard(type);
    }
    
    // 진행도 업데이트
    async updateProgression() {
        try {
            const response = await fetch(`${this.services.progression}/player/${this.playerId}`);
            const data = await response.json();
            
            if (data.level) {
                this.playerData.level = data.level;
                this.playerData.xp = data.xp;
                this.playerData.achievements = data.achievements || [];
            }
            
            this.updateXPDisplay();
            this.updateAchievementsGrid();
        } catch (error) {
            console.error('진행도 로드 실패:', error);
        }
    }
    
    updateAchievementsGrid() {
        const grid = document.getElementById('achievementsGrid');
        const allAchievements = [
            { id: 'level_5', icon: 'fas fa-star', name: '레벨 5', desc: '레벨 5 달성' },
            { id: 'clicker_100', icon: 'fas fa-mouse', name: '클릭 마스터', desc: '100번 클릭' },
            { id: 'memory_master', icon: 'fas fa-brain', name: '기억의 달인', desc: '메모리 레벨 10' },
            { id: 'speed_demon', icon: 'fas fa-bolt', name: '번개 손', desc: '200ms 반응속도' }
        ];
        
        grid.innerHTML = allAchievements.map(achievement => {
            const unlocked = this.playerData.achievements.includes(achievement.id);
            return `
                <div class="achievement-badge ${unlocked ? 'unlocked' : ''}">
                    <i class="${achievement.icon}"></i>
                    <div class="achievement-name">${achievement.name}</div>
                    <div class="achievement-desc">${achievement.desc}</div>
                </div>
            `;
        }).join('');
    }
    
    // 모니터링 업데이트
    async updateMonitoring() {
        await this.checkServiceHealth();
        await this.updateFairnessStatus();
    }
    
    async checkServiceHealth() {
        const services = ['gateway', 'leaderboard', 'progression', 'fairness'];
        
        for (const service of services) {
            try {
                const response = await fetch(`${this.services[service]}/healthz`, { 
                    method: 'GET',
                    timeout: 5000 
                });
                
                const statusEl = document.getElementById(`${service}Status`) || 
                               document.getElementById(`${service}ServiceStatus`);
                
                if (statusEl) {
                    if (response.ok) {
                        statusEl.innerHTML = '<div class="status-dot online"></div><span>온라인</span>';
                    } else {
                        statusEl.innerHTML = '<div class="status-dot offline"></div><span>오프라인</span>';
                    }
                }
            } catch (error) {
                const statusEl = document.getElementById(`${service}Status`) || 
                               document.getElementById(`${service}ServiceStatus`);
                if (statusEl) {
                    statusEl.innerHTML = '<div class="status-dot offline"></div><span>연결 실패</span>';
                }
            }
        }
    }
    
    async updateFairnessStatus() {
        try {
            // 공정성 상태는 실제로는 서비스에서 가져와야 하지만, 
            // 데모용으로 정상 상태로 표시
            document.getElementById('personalFairnessStatus').innerHTML = 
                '<div class="status-indicator green"></div><span>정상 플레이</span>';
            
            document.getElementById('eventFrequency').textContent = '정상';
            document.getElementById('scorePattern').textContent = '정상';
            document.getElementById('lastFairnessCheck').textContent = new Date().toLocaleString();
            
        } catch (error) {
            console.error('공정성 상태 업데이트 실패:', error);
        }
    }
    
    // 이벤트 전송
    async sendGameEvent(eventType, payload) {
        try {
            const event = {
                type: eventType,
                playerId: this.playerId,
                ts: Date.now(),
                payload: payload
            };
            
            await fetch(`${this.services.gateway}/events`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(event)
            });
            
            this.addEventLog(`게임 이벤트: ${eventType}`);
        } catch (error) {
            console.error('이벤트 전송 실패:', error);
            this.addEventLog(`이벤트 전송 실패: ${eventType}`, 'error');
        }
    }
    
    async sendProgressionEvent(eventType, payload) {
        try {
            const event = {
                type: eventType,
                playerId: this.playerId,
                ts: Date.now(),
                payload: payload
            };
            
            await fetch(`${this.services.progression}/xp`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(event)
            });
        } catch (error) {
            console.error('진행도 이벤트 전송 실패:', error);
        }
    }
    
    // 플레이어 진행도 로드
    async loadPlayerProgress() {
        try {
            const response = await fetch(`${this.services.progression}/player/${this.playerId}`);
            const data = await response.json();
            
            if (data.level) {
                this.playerData = { ...this.playerData, ...data };
            }
        } catch (error) {
            console.log('진행도 로드 실패 (새 플레이어):', error);
        }
    }
    
    // 주기적 업데이트
    startPeriodicUpdates() {
        // 리더보드 5초마다 업데이트
        setInterval(() => {
            if (this.currentTab === 'leaderboard') {
                this.loadLeaderboard();
            }
        }, 5000);
        
        // 대시보드 10초마다 업데이트
        setInterval(() => {
            if (this.currentTab === 'dashboard') {
                this.updateDashboard();
            }
        }, 10000);
        
        // 서비스 상태 30초마다 체크
        setInterval(() => {
            if (this.currentTab === 'admin') {
                this.checkServiceHealth();
            }
        }, 30000);
    }
    
    // 유틸리티 함수들
    sleep(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }
    
    showLoading(elementId) {
        document.getElementById(elementId).innerHTML = 
            '<div class="loading"><i class="fas fa-spinner fa-spin"></i> 로딩 중...</div>';
    }
    
    showNotification(message, type = 'info') {
        const notifications = document.getElementById('notifications');
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        notification.innerHTML = `
            <div><strong>${type === 'success' ? '성공' : type === 'error' ? '오류' : '알림'}</strong></div>
            <div>${message}</div>
        `;
        
        notifications.appendChild(notification);
        
        setTimeout(() => {
            notification.remove();
        }, 5000);
    }
    
    showXPGain(amount) {
        const xpElement = document.getElementById('playerXP');
        const popup = document.createElement('div');
        popup.style.position = 'absolute';
        popup.style.color = 'var(--success-color)';
        popup.style.fontWeight = 'bold';
        popup.style.animation = 'fadeUp 2s ease-out forwards';
        popup.textContent = `+${amount} XP`;
        
        const rect = xpElement.getBoundingClientRect();
        popup.style.left = rect.right + 'px';
        popup.style.top = rect.top + 'px';
        popup.style.zIndex = '1000';
        
        document.body.appendChild(popup);
        setTimeout(() => popup.remove(), 2000);
    }
    
    showAchievement(achievementId) {
        const achievements = {
            'level_5': '레벨 5 달성!',
            'clicker_100': '100번 클릭 달성!',
            'memory_master': '메모리 마스터!',
            'speed_demon': '번개 손!'
        };
        
        this.showNotification(`🏆 새 업적: ${achievements[achievementId]}`, 'success');
    }
    
    addEventLog(message, type = 'info') {
        const log = document.getElementById('eventLog');
        const item = document.createElement('div');
        item.className = `log-item ${type}`;
        item.innerHTML = `${new Date().toLocaleTimeString()} - ${message}`;
        
        log.insertBefore(item, log.firstChild);
        
        // 로그 항목이 너무 많으면 제거
        while (log.children.length > 50) {
            log.removeChild(log.lastChild);
        }
    }
    
    claimAllRewards() {
        this.showNotification('보상을 받았습니다!', 'success');
        document.getElementById('claimAllRewards').classList.add('hidden');
        document.getElementById('pendingRewards').innerHTML = 
            '<div class="no-rewards">받을 수 있는 보상이 없습니다</div>';
    }
}

// CSS 애니메이션 추가
const style = document.createElement('style');
style.textContent = `
    @keyframes fadeUp {
        0% { opacity: 1; transform: translateY(0); }
        100% { opacity: 0; transform: translateY(-30px); }
    }
`;
document.head.appendChild(style);

// 페이지 로드 시 플랫폼 초기화
document.addEventListener('DOMContentLoaded', () => {
    window.liveOpsPlatform = new LiveOpsPlatform();
});
