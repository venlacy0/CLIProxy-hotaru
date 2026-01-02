package donation

// HTML templates for donation site pages

const indexPageHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>捐赠站点</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 16px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            padding: 40px;
            max-width: 420px;
            width: 100%;
            text-align: center;
        }
        .logo { font-size: 48px; margin-bottom: 20px; }
        h1 { color: #333; margin-bottom: 10px; font-size: 24px; }
        .subtitle { color: #666; margin-bottom: 30px; font-size: 14px; }
        .btn {
            display: inline-block;
            padding: 14px 32px;
            border-radius: 8px;
            font-size: 16px;
            font-weight: 600;
            text-decoration: none;
            cursor: pointer;
            border: none;
            transition: all 0.3s;
            width: 100%;
            margin-bottom: 12px;
        }
        .btn-primary {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
        }
        .btn-primary:hover { transform: translateY(-2px); box-shadow: 0 8px 20px rgba(102,126,234,0.4); }
        .btn-secondary {
            background: #f5f5f5;
            color: #333;
        }
        .btn-secondary:hover { background: #eee; }
        .user-info {
            background: #f8f9fa;
            border-radius: 12px;
            padding: 20px;
            margin-bottom: 24px;
            text-align: left;
        }
        .user-info h3 { color: #333; margin-bottom: 12px; font-size: 16px; }
        .user-info p { color: #666; font-size: 14px; margin-bottom: 8px; }
        .user-info .role {
            display: inline-block;
            padding: 4px 12px;
            border-radius: 20px;
            font-size: 12px;
            font-weight: 600;
        }
        .role-admin { background: #ffd700; color: #333; }
        .role-user { background: #e0e0e0; color: #666; }
        .status-bound { color: #28a745; }
        .status-unbound { color: #dc3545; }
        .hidden { display: none; }
        .loading { opacity: 0.6; pointer-events: none; }
        .message {
            padding: 12px 16px;
            border-radius: 8px;
            margin-bottom: 16px;
            font-size: 14px;
        }
        .message-success { background: #d4edda; color: #155724; }
        .message-error { background: #f8d7da; color: #721c24; }
        .input-group { margin-bottom: 16px; text-align: left; }
        .input-group label { display: block; margin-bottom: 6px; color: #333; font-size: 14px; font-weight: 500; }
        .input-group input {
            width: 100%;
            padding: 12px 16px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            font-size: 16px;
            transition: border-color 0.3s;
        }
        .input-group input:focus { outline: none; border-color: #667eea; }
        .quota-info {
            background: linear-gradient(135deg, #28a745 0%, #20c997 100%);
            color: white;
            padding: 20px;
            border-radius: 12px;
            margin-bottom: 24px;
        }
        .quota-info h3 { font-size: 14px; opacity: 0.9; margin-bottom: 8px; }
        .quota-info .amount { font-size: 36px; font-weight: 700; }
        .footer { margin-top: 24px; color: #999; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <!-- 未登录状态 -->
        <div id="guest-view">
            <div class="logo">🎁</div>
            <h1>欢迎来到捐赠站点</h1>
            <p class="subtitle">通过 Linux Do 账号登录，绑定您的 API 账户</p>
            <a href="/linuxdo/login" class="btn btn-primary">
                🔐 使用 Linux Do 登录
            </a>
        </div>

        <!-- 已登录状态 -->
        <div id="user-view" class="hidden">
            <div class="user-info">
                <h3>👤 用户信息</h3>
                <p>用户名: <strong id="username"></strong></p>
                <p>Linux Do ID: <strong id="linux-do-id"></strong></p>
                <p>角色: <span id="role" class="role"></span></p>
                <p>绑定状态: <span id="bind-status"></span></p>
                <p id="newapi-id-row" class="hidden">New-API ID: <strong id="newapi-id"></strong></p>
            </div>

            <div id="message" class="message hidden"></div>

            <!-- 未绑定：显示绑定表单 -->
            <div id="bind-form" class="hidden">
                <div class="input-group">
                    <label for="newapi-user-id">请输入您的 New-API 用户 ID</label>
                    <input type="number" id="newapi-user-id" placeholder="例如: 12345">
                </div>
                <button onclick="submitBind()" class="btn btn-primary">绑定账户</button>
            </div>

            <!-- 已绑定：显示捐赠信息 -->
            <div id="donate-view" class="hidden">
                <div class="quota-info">
                    <h3>捐赠奖励额度</h3>
                    <div class="amount">$<span id="quota-amount">20</span></div>
                </div>
                <button onclick="confirmDonate()" class="btn btn-primary">✨ 确认捐赠</button>
            </div>

            <button onclick="logout()" class="btn btn-secondary">退出登录</button>
        </div>

        <div class="footer">
            Powered by CLI Proxy API
        </div>
    </div>

    <script>
        async function checkStatus() {
            try {
                const resp = await fetch('/status');
                const data = await resp.json();
                
                if (data.logged_in) {
                    document.getElementById('guest-view').classList.add('hidden');
                    document.getElementById('user-view').classList.remove('hidden');
                    
                    document.getElementById('username').textContent = data.user.username;
                    document.getElementById('linux-do-id').textContent = data.user.linux_do_id;
                    
                    const roleEl = document.getElementById('role');
                    roleEl.textContent = data.user.role === 'admin' ? '管理员' : '普通用户';
                    roleEl.className = 'role ' + (data.user.role === 'admin' ? 'role-admin' : 'role-user');
                    
                    const bindStatus = document.getElementById('bind-status');
                    if (data.bound) {
                        bindStatus.textContent = '已绑定';
                        bindStatus.className = 'status-bound';
                        document.getElementById('newapi-id-row').classList.remove('hidden');
                        document.getElementById('newapi-id').textContent = data.user.newapi_user_id;
                        document.getElementById('bind-form').classList.add('hidden');
                        document.getElementById('donate-view').classList.remove('hidden');
                        loadDonateInfo();
                    } else {
                        bindStatus.textContent = '未绑定';
                        bindStatus.className = 'status-unbound';
                        document.getElementById('bind-form').classList.remove('hidden');
                        document.getElementById('donate-view').classList.add('hidden');
                    }
                } else {
                    document.getElementById('guest-view').classList.remove('hidden');
                    document.getElementById('user-view').classList.add('hidden');
                }
            } catch (e) {
                console.error('Failed to check status:', e);
            }
        }

        async function loadDonateInfo() {
            try {
                const resp = await fetch('/donate');
                const data = await resp.json();
                if (data.quota_amount) {
                    document.getElementById('quota-amount').textContent = (data.quota_amount / 100000).toFixed(0);
                }
            } catch (e) {
                console.error('Failed to load donate info:', e);
            }
        }

        function showMessage(text, isError) {
            const msg = document.getElementById('message');
            msg.textContent = text;
            msg.className = 'message ' + (isError ? 'message-error' : 'message-success');
            msg.classList.remove('hidden');
            setTimeout(() => msg.classList.add('hidden'), 5000);
        }

        async function submitBind() {
            const userIdInput = document.getElementById('newapi-user-id');
            const userId = parseInt(userIdInput.value);
            if (!userId || userId <= 0) {
                showMessage('请输入有效的用户 ID', true);
                return;
            }

            try {
                const resp = await fetch('/bind', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ newapi_user_id: userId })
                });
                const data = await resp.json();
                
                if (resp.ok) {
                    showMessage('绑定成功！', false);
                    setTimeout(() => checkStatus(), 1000);
                } else {
                    showMessage(data.message || '绑定失败', true);
                }
            } catch (e) {
                showMessage('网络错误，请重试', true);
            }
        }

        async function confirmDonate() {
            if (!confirm('确认已完成捐赠？系统将为您添加额度。')) return;

            try {
                const resp = await fetch('/donate/confirm', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' }
                });
                const data = await resp.json();
                
                if (resp.ok) {
                    showMessage('🎉 捐赠成功！额度已添加到您的账户', false);
                } else {
                    showMessage(data.message || '操作失败', true);
                }
            } catch (e) {
                showMessage('网络错误，请重试', true);
            }
        }

        async function logout() {
            try {
                await fetch('/logout', { method: 'POST' });
                window.location.reload();
            } catch (e) {
                window.location.reload();
            }
        }

        // 页面加载时检查状态
        checkStatus();
    </script>
</body>
</html>`

// GetIndexPageHTML returns the index page HTML template.
func GetIndexPageHTML() string {
	return indexPageHTML
}
