# Implementation Plan: Donation Site

## Overview

本实现计划将捐赠站点功能分解为可执行的编码任务。实现顺序遵循依赖关系：先配置和数据模型，再核心服务，然后中间件和处理器，最后集成测试。

## Tasks

- [x] 1. 扩展配置结构支持Linux Do Connect和捐赠设置
  - [x] 1.1 在config.go中添加LinuxDoConnectConfig结构体
    - 添加ClientID、ClientSecret、RedirectURI、AuthURL、TokenURL、UserInfoURL字段
    - 添加默认URL值（AuthURL: https://connect.linux.do/oauth2/authorize, TokenURL: https://connect.linux.do/oauth2/token, UserInfoURL: https://connect.linux.do/api/user）
    - _Requirements: 2.1, 2.2, 2.3_
  - [x] 1.2 在config.go中添加DonationConfig结构体
    - 添加QuotaAmount字段（默认2000000，即$20）
    - 添加AdminLinuxDoIDs列表（[]int类型）
    - _Requirements: 8.3_
  - [x] 1.3 更新Config结构体包含新配置
    - 添加LinuxDoConnect LinuxDoConnectConfig字段
    - 添加Donation DonationConfig字段
    - _Requirements: 2.1, 2.2, 2.3_
  - [ ]* 1.4 编写配置解析属性测试
    - **Property 2: 配置文件解析正确性**
    - **Validates: Requirements 2.1, 2.2, 2.3**

- [x] 2. 实现数据模型和存储服务
  - [x] 2.1 创建internal/donation/models.go定义数据结构
    - 定义LinuxDoUser结构体（ID int, Username string, Name string, Email string, Avatar string）
    - 定义Session结构体（ID string, LinuxDoID int, Username string, Role string, NewAPIUserID int, CreatedAt time.Time, ExpiresAt time.Time）
    - 定义UserBinding结构体（LinuxDoID int, NewAPIUserID int, BoundAt time.Time）
    - _Requirements: 7.1, 3.7_
  - [x] 2.2 创建internal/donation/session_store.go实现会话存储
    - 实现SessionStore结构体（使用sync.Map存储）
    - 实现NewSessionStore构造函数
    - 实现Create方法（创建会话，设置24小时过期）
    - 实现Get方法（获取会话，检查过期）
    - 实现Update方法（更新会话）
    - 实现Delete方法（删除会话）
    - 实现GenerateToken方法（生成32字节随机令牌，返回hex编码）
    - _Requirements: 7.1, 7.6_
  - [ ]* 2.3 编写会话令牌生成属性测试
    - **Property 7: 会话令牌安全性**
    - **Validates: Requirements 7.1**
  - [x] 2.4 创建internal/donation/binding_store.go实现绑定存储
    - 实现BindingStore结构体（JSON文件存储，路径为auth-dir/bindings.json）
    - 实现NewBindingStore构造函数
    - 实现GetByLinuxDoID方法
    - 实现Create方法
    - 实现Delete方法
    - 实现load和save内部方法
    - _Requirements: 3.7_
  - [ ]* 2.5 编写绑定数据往返属性测试
    - **Property 4: 绑定数据持久化往返一致性**
    - **Validates: Requirements 3.7**

- [x] 3. Checkpoint - 确保数据层测试通过
  - 确保所有测试通过，如有问题请询问用户

- [x] 4. 实现Linux Do Connect OAuth服务
  - [x] 4.1 创建internal/donation/linuxdo_service.go
    - 实现LinuxDoConnectService结构体
    - 实现NewLinuxDoConnectService构造函数（接收LinuxDoConnectConfig）
    - 实现GenerateAuthURL方法（生成OAuth授权URL，包含client_id、redirect_uri、response_type=code、state参数）
    - 实现ExchangeToken方法（用授权码交换访问令牌）
    - 实现GetUserInfo方法（使用访问令牌获取用户信息）
    - _Requirements: 1.2, 1.3, 1.4_
  - [ ]* 4.2 编写OAuth URL生成属性测试
    - **Property 1: OAuth重定向URL格式正确性**
    - **Validates: Requirements 1.2**

- [x] 5. 实现New-API服务
  - [x] 5.1 创建internal/donation/newapi_service.go
    - 实现NewAPIService结构体
    - 实现NewNewAPIService构造函数（从环境变量读取NEW_API_BASE_URL和NEW_API_ADMIN_TOKEN）
    - 实现GetUserByID方法（调用New-API获取用户信息，返回包含linux_do_id的用户数据）
    - 实现AddQuota方法（调用New-API为用户增加额度）
    - 所有请求使用Bearer token认证
    - _Requirements: 4.1, 4.2, 4.4, 5.2_
  - [ ]* 5.2 编写New-API请求认证头属性测试
    - **Property 11: New-API请求认证头**
    - **Validates: Requirements 4.4**

- [x] 6. 实现认证和角色中间件
  - [x] 6.1 创建internal/donation/middleware.go
    - 实现AuthMiddleware函数（从Cookie读取session_id，验证会话有效性，将会话存入gin.Context）
    - 实现RoleMiddleware函数（从Context获取会话，验证角色权限）
    - 无效会话返回401，权限不足返回403
    - _Requirements: 7.3, 7.4, 6.1-6.5_
  - [ ]* 6.2 编写会话验证属性测试
    - **Property 8: 会话验证正确性**
    - **Validates: Requirements 7.3, 7.4**
  - [ ]* 6.3 编写访问控制属性测试
    - **Property 5: 普通用户Auth_File访问控制**
    - **Property 6: 管理员Auth_File访问权限**
    - **Validates: Requirements 6.1, 6.2, 6.3, 6.4, 6.5**

- [x] 7. 实现角色分配逻辑
  - [x] 7.1 创建internal/donation/role_service.go
    - 实现RoleService结构体
    - 实现NewRoleService构造函数（接收管理员ID列表）
    - 实现DetermineRole方法（根据LinuxDoID判断角色，在列表中返回"admin"，否则返回"user"）
    - _Requirements: 8.2, 8.4_
  - [ ]* 7.2 编写角色分配属性测试
    - **Property 9: 默认用户角色分配**
    - **Property 10: 管理员角色分配**
    - **Validates: Requirements 8.2, 8.4**

- [x] 8. Checkpoint - 确保服务层测试通过
  - 确保所有测试通过，如有问题请询问用户

- [x] 9. 实现HTTP处理器
  - [x] 9.1 创建internal/donation/handlers.go - 登录处理器
    - 实现DonationHandler结构体（包含所有服务依赖）
    - 实现NewDonationHandler构造函数
    - 实现HandleLogin方法（GET /linuxdo/login，生成授权URL并重定向）
    - 实现HandleCallback方法（GET /linuxdo/callback，处理OAuth回调，创建会话，设置Cookie）
    - 实现HandleLogout方法（POST /logout，销毁会话，清除Cookie）
    - _Requirements: 1.1, 1.2, 1.5, 1.6, 7.5_
  - [x] 9.2 添加用户绑定处理器
    - 实现HandleBindPage方法（GET /bind，返回绑定状态JSON）
    - 实现HandleBind方法（POST /bind，验证New-API用户ID，检查linux_do_id匹配，保存绑定）
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6_
  - [ ]* 9.3 编写用户ID匹配验证属性测试
    - **Property 3: 用户ID匹配验证正确性**
    - **Validates: Requirements 3.4, 3.5**
  - [x] 9.4 添加捐赠处理器
    - 实现HandleDonatePage方法（GET /donate，返回捐赠状态JSON）
    - 实现HandleDonateConfirm方法（POST /donate/confirm，调用New-API增加额度，记录日志）
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_

- [x] 10. 实现日志服务
  - [x] 10.1 创建internal/donation/logger.go
    - 实现DonationLogger结构体
    - 实现NewDonationLogger构造函数
    - 实现LogDonation方法（记录时间戳、用户ID、额度变更）
    - 实现LogError方法（记录错误详情）
    - 实现敏感信息过滤（过滤admin_token、client_secret等）
    - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5_
  - [ ]* 10.2 编写日志属性测试
    - **Property 12: 捐赠日志完整性**
    - **Property 13: 日志敏感信息过滤**
    - **Validates: Requirements 5.5, 9.5**

- [x] 11. 集成路由和修改Auth_File访问控制
  - [x] 11.1 修改internal/api/server.go注册捐赠路由
    - 创建/linuxdo路由组（login、callback）
    - 创建需要认证的路由组（/bind、/donate、/logout）
    - 应用AuthMiddleware到需要认证的路由
    - _Requirements: 1.1_
  - [x] 11.2 修改internal/api/handlers/management/auth_files.go添加角色检查
    - 在ListAuthFiles方法开头添加角色验证
    - 在DownloadAuthFile方法开头添加角色验证
    - 在UploadAuthFile方法开头添加角色验证
    - 在DeleteAuthFile方法开头添加角色验证
    - 普通用户（role != "admin"）返回403 Forbidden
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 6.6_

- [x] 12. 更新配置示例文件
  - [x] 12.1 更新config.example.yaml
    - 添加linux-do-connect配置示例（client-id、client-secret、redirect-uri）
    - 添加admin-linux-do-ids配置示例
    - 添加donation配置示例（quota-amount）
    - _Requirements: 2.1, 2.2, 2.3_
  - [x] 12.2 更新.env.example
    - 添加NEW_API_ADMIN_TOKEN示例
    - 添加NEW_API_BASE_URL示例
    - _Requirements: 4.1, 4.2_

- [x] 13. Final Checkpoint - 确保所有测试通过
  - 确保所有测试通过，如有问题请询问用户

## Notes

- 带 `*` 标记的任务为可选测试任务，可跳过以加快MVP开发
- 每个任务引用了具体的需求编号以便追溯
- 检查点任务用于确保增量验证
- 属性测试验证通用正确性属性
- 单元测试验证具体示例和边界情况
- 所有新代码放在internal/donation目录下，保持与现有代码结构一致
