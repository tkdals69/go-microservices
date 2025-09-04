class ClickGame {
    constructor() {
        this.score = 0;
        this.multiplier = 1;
        this.playerId = null; // IP 기반 플레이어 ID
        this.clickBtn = document.getElementById('clickBtn');
        this.scoreValue = document.getElementById('scoreValue');
        this.multiplierValue = document.getElementById('multiplierValue');
        this.status = document.getElementById('status');
        this.leaderboardList = document.getElementById('leaderboardList');
        
        // 백엔드 URL 설정 (환경에 따라 동적 설정)
        this.backendUrl = this.getBackendUrl();
        
        // 게임 초기화
        this.init();
    }

    getBackendUrl() {
        // 현재 호스트가 localhost가 아니면 같은 도메인의 8085 포트 사용
        if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
            return 'http://localhost:8085';
        } else {
            // Azure 환경에서는 Gateway 서비스의 FQDN 사용
            return `http://gomicroservices-gateway.koreacentral.azurecontainer.io:8085`;
        }
    }

    async init() {
        // 서버에서 클라이언트 IP를 가져와서 플레이어 ID로 사용
        await this.initializePlayerId();
        
        this.clickBtn.addEventListener('click', () => this.handleClick());
        
        // 10초마다 리더보드 갱신
        this.updateLeaderboard();
        setInterval(() => this.updateLeaderboard(), 10000);
    }

    async initializePlayerId() {
        try {
            const response = await fetch(`${this.backendUrl}/client-ip`);
            const data = await response.json();
            this.playerId = data.ip;
            this.status.textContent = `플레이어 ID: ${this.playerId}`;
        } catch (error) {
            console.error('플레이어 ID 초기화 실패:', error);
            this.playerId = 'unknown-' + Date.now(); // 폴백으로 타임스탬프 사용
            this.status.textContent = `플레이어 ID: ${this.playerId} (오프라인)`;
        }
    }

    async handleClick() {
        // 점수 증가
        const points = 1 * this.multiplier;
        this.score += points;
        this.scoreValue.textContent = this.score;

        // 100점마다 배율 증가
        if (this.score % 100 === 0) {
            this.multiplier++;
            this.multiplierValue.textContent = this.multiplier;
        }

        // 백엔드로 이벤트 전송
        await this.sendEvent(points);
    }

    async sendEvent(points) {
        try {
            const event = {
                type: "progression",
                playerId: this.playerId, // IP 기반 플레이어 ID 사용
                ts: Math.floor(Date.now() / 1000),
                payload: {
                    deltaXp: points
                }
            };

            // HMAC 서명 생성 (실제 구현시 서버사이드에서 처리해야 함)
            const sig = await this.generateHMAC(JSON.stringify(event));

            const response = await fetch(`${this.backendUrl}/events`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-Signature': `sha256=${sig}`,
                    'Idempotency-Key': `click-${Date.now()}`
                },
                body: JSON.stringify(event)
            });

            if (response.status === 202) {
                this.status.textContent = `점수 전송 성공! (${this.playerId})`;
            } else {
                throw new Error(`서버 응답 ${response.status}`);
            }
        } catch (error) {
            this.status.textContent = `오류: ${error.message}`;
            console.error('이벤트 전송 실패:', error);
        }
    }

    async updateLeaderboard() {
        try {
            const response = await fetch(`${this.backendUrl}/leaderboard?window=daily&limit=10`);
            const data = await response.json();
            
            this.leaderboardList.innerHTML = data.map((entry, index) => {
                const playerDisplay = entry.playerId === this.playerId 
                    ? `${entry.playerId} (나)` 
                    : entry.playerId;
                return `
                <div class="leaderboard-entry ${entry.playerId === this.playerId ? 'current-player' : ''}">
                    ${index + 1}. ${playerDisplay}: ${entry.score}점
                </div>`;
            }).join('');
        } catch (error) {
            console.error('리더보드 갱신 실패:', error);
        }
    }

    // 실제 구현시 서버사이드에서 처리해야 함
    async generateHMAC(message) {
        const encoder = new TextEncoder();
        const key = encoder.encode('your-hmac-secret'); // .env의 HMAC_SECRET과 동일하게 설정
        const data = encoder.encode(message);
        
        const cryptoKey = await crypto.subtle.importKey(
            'raw',
            key,
            { name: 'HMAC', hash: 'SHA-256' },
            false,
            ['sign']
        );
        
        const signature = await crypto.subtle.sign(
            'HMAC',
            cryptoKey,
            data
        );

        return Array.from(new Uint8Array(signature))
            .map(b => b.toString(16).padStart(2, '0'))
            .join('');
    }
}

// 게임 인스턴스 생성
new ClickGame();
