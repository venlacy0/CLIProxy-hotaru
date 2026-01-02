# Requirements Document

## Introduction

本文档定义了基于CLIProxyAPI项目改造的捐赠站点系统需求。该系统允许用户通过Linux Do Connect登录，绑定new-api账户ID，并在成功捐赠后获得$20额度奖励。

## Glossary

- **Donation_Site**: 捐赠站点系统，基于CLIProxyAPI改造
- **Linux_Do_Connect**: Linux Do社区的OAuth认证服务
- **New_API**: 用户的API服务平台，提供用户管理和额度管理功能
- **User_Linux_Do_ID**: 用户在Linux Do平台的唯一标识符
- **New_API_User_ID**: 用户在new-api平台的用户ID
- **Quota**: new-api平台中用户的可用额度
- **Auth_File**: 系统中存储认证信息的文件
- **Admin_Token**: new-api管理员API令牌，用于调用管理接口

## Requirements

### Requirement 1: Linux Do Connect OAuth登录

**User Story:** 作为用户，我希望通过Linux Do Connect登录捐赠站点，以便使用我的Linux Do身份进行认证。

#### Acceptance Criteria

1. WHEN 用户访问登录页面, THE Donation_Site SHALL 显示Linux Do Connect登录按钮
2. WHEN 用户点击登录按钮, THE Donation_Site SHALL 重定向用户到Linux Do Connect授权页面
3. WHEN Linux Do Connect返回授权码, THE Donation_Site SHALL 使用授权码交换访问令牌
4. WHEN 访问令牌获取成功, THE Donation_Site SHALL 获取用户的User_Linux_Do_ID和基本信息
5. WHEN 登录成功, THE Donation_Site SHALL 创建用户会话并重定向到主页
6. IF 授权失败或被拒绝, THEN THE Donation_Site SHALL 显示错误信息并允许重试

### Requirement 2: Linux Do Connect配置管理

**User Story:** 作为管理员，我希望在配置文件中设置Linux Do Connect的凭据，以便系统能够进行OAuth认证。

#### Acceptance Criteria

1. THE Donation_Site SHALL 从配置文件读取Linux Do Connect的client_id
2. THE Donation_Site SHALL 从配置文件读取Linux Do Connect的client_secret
3. THE Donation_Site SHALL 从配置文件读取Linux Do Connect的redirect_uri
4. WHEN 配置文件缺少必要的Connect凭据, THE Donation_Site SHALL 在启动时记录错误日志
5. THE Donation_Site SHALL 支持配置文件热重载以更新Connect凭据

### Requirement 3: New-API用户ID绑定

**User Story:** 作为用户，我希望绑定我的new-api用户ID，以便系统能够验证我的身份并在捐赠后给予额度。

#### Acceptance Criteria

1. WHEN 用户首次登录且未绑定New_API_User_ID, THE Donation_Site SHALL 显示ID绑定页面
2. WHEN 用户提交New_API_User_ID, THE Donation_Site SHALL 调用New_API接口验证该ID是否存在
3. WHEN 用户提交New_API_User_ID, THE Donation_Site SHALL 调用New_API接口获取该用户的User_Linux_Do_ID
4. WHEN New_API返回的User_Linux_Do_ID与当前登录用户匹配, THE Donation_Site SHALL 完成绑定并保存关联
5. IF New_API返回的User_Linux_Do_ID与当前登录用户不匹配, THEN THE Donation_Site SHALL 拒绝绑定并显示错误信息
6. IF New_API_User_ID不存在, THEN THE Donation_Site SHALL 显示"用户不存在"错误
7. WHEN 绑定成功, THE Donation_Site SHALL 将绑定信息持久化存储

### Requirement 4: New-API管理员令牌配置

**User Story:** 作为管理员，我希望通过环境变量配置new-api管理员令牌，以便系统能够调用new-api的管理接口。

#### Acceptance Criteria

1. THE Donation_Site SHALL 从环境变量NEW_API_ADMIN_TOKEN读取管理员令牌
2. THE Donation_Site SHALL 从环境变量NEW_API_BASE_URL读取new-api的基础URL
3. WHEN 环境变量未设置, THE Donation_Site SHALL 在启动时记录警告日志
4. WHEN 调用New_API管理接口, THE Donation_Site SHALL 使用Admin_Token进行认证
5. IF Admin_Token无效或过期, THEN THE Donation_Site SHALL 记录错误并返回适当的错误响应

### Requirement 5: 捐赠流程处理

**User Story:** 作为用户，我希望完成捐赠后自动获得$20额度，以便我能够使用new-api的服务。

#### Acceptance Criteria

1. WHEN 用户已绑定New_API_User_ID且访问捐赠页面, THE Donation_Site SHALL 显示捐赠入口
2. WHEN 捐赠成功确认, THE Donation_Site SHALL 调用New_API接口为用户增加$20额度
3. WHEN 额度增加成功, THE Donation_Site SHALL 显示成功消息并记录捐赠日志
4. IF 额度增加失败, THEN THE Donation_Site SHALL 显示错误信息并记录失败日志以便人工处理
5. THE Donation_Site SHALL 记录每次捐赠的时间、用户ID和额度变更

### Requirement 6: 认证文件访问控制

**User Story:** 作为管理员，我希望普通用户无法查看认证文件，以保护系统安全。

#### Acceptance Criteria

1. WHEN 普通用户请求访问Auth_File列表接口, THE Donation_Site SHALL 返回403禁止访问
2. WHEN 普通用户请求下载Auth_File, THE Donation_Site SHALL 返回403禁止访问
3. WHEN 普通用户请求上传Auth_File, THE Donation_Site SHALL 返回403禁止访问
4. WHEN 普通用户请求删除Auth_File, THE Donation_Site SHALL 返回403禁止访问
5. WHEN 管理员请求访问Auth_File相关接口, THE Donation_Site SHALL 正常处理请求
6. THE Donation_Site SHALL 在后端验证用户角色，不依赖前端隐藏

### Requirement 7: 用户会话管理

**User Story:** 作为用户，我希望系统能够安全地管理我的登录会话，以保护我的账户安全。

#### Acceptance Criteria

1. WHEN 用户登录成功, THE Donation_Site SHALL 生成安全的会话令牌
2. THE Donation_Site SHALL 将会话令牌存储在HTTP-only Cookie中
3. WHEN 用户请求需要认证的接口, THE Donation_Site SHALL 验证会话令牌有效性
4. WHEN 会话令牌无效或过期, THE Donation_Site SHALL 返回401未授权并清除Cookie
5. WHEN 用户主动登出, THE Donation_Site SHALL 销毁会话并清除Cookie
6. THE Donation_Site SHALL 设置会话过期时间为24小时

### Requirement 8: 用户角色区分

**User Story:** 作为系统，我需要区分普通用户和管理员，以便实施不同的访问控制策略。

#### Acceptance Criteria

1. THE Donation_Site SHALL 支持两种用户角色：普通用户和管理员
2. WHEN 用户通过Linux Do Connect登录, THE Donation_Site SHALL 默认分配普通用户角色
3. THE Donation_Site SHALL 从配置文件读取管理员User_Linux_Do_ID列表
4. WHEN 登录用户的User_Linux_Do_ID在管理员列表中, THE Donation_Site SHALL 分配管理员角色
5. THE Donation_Site SHALL 在会话中存储用户角色信息

### Requirement 9: 错误处理与日志

**User Story:** 作为管理员，我希望系统能够记录详细的操作日志和错误信息，以便排查问题。

#### Acceptance Criteria

1. WHEN 发生OAuth认证错误, THE Donation_Site SHALL 记录错误详情到日志
2. WHEN 发生New_API调用错误, THE Donation_Site SHALL 记录请求和响应详情到日志
3. WHEN 发生用户绑定操作, THE Donation_Site SHALL 记录操作详情到日志
4. WHEN 发生捐赠操作, THE Donation_Site SHALL 记录操作详情到日志
5. THE Donation_Site SHALL 不在日志中记录敏感信息如令牌和密钥

