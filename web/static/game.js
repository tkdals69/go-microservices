class ClickGame {
    constructor() {
        this.score = 0;
        this.multiplier = 1;
        this.clickBtn = document.getElementById('clickBtn');
        this.scoreValue = document.getElementById('scoreValue');
        this.multiplierValue = document.getElementById('multiplierValue');
        this.status = document.getElementById('status');
        this.leaderboardList = document.getElementById('leaderboardList');
        
        // 게임 초기화
        this.init();
    }

    async init() {
        this.clickBtn.addEventListener('click', () => this.handleClick());
        
        // 10초마다 리더보드 갱신
        this.updateLeaderboard();
        setInterval(() => this.updateLeaderboard(), 10000);
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
                playerId: "player1", // 실제 구현시 로그인한 플레이어 ID 사용
                ts: Math.floor(Date.now() / 1000),
                payload: {
                    deltaXp: points
                }
            };

            // HMAC 서명 생성 (실제 구현시 서버사이드에서 처리해야 함)
            const sig = await this.generateHMAC(JSON.stringify(event));

            const response = await fetch('http://localhost:8080/events', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-Signature': `sha256=${sig}`,
                    'Idempotency-Key': `click-${Date.now()}`
                },
                body: JSON.stringify(event)
            });

            if (response.status === 202) {
                this.status.textContent = '점수 전송 성공!';
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
            const response = await fetch('http://localhost:8080/leaderboard?window=daily&limit=10');
            const data = await response.json();
            
            this.leaderboardList.innerHTML = data.map((entry, index) => `
                <div class="leaderboard-entry">
                    ${index + 1}. ${entry.playerId}: ${entry.score}점
                </div>
            `).join('');
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
