# systemd 最小示例

## 安装
1. 复制二进制到 `/usr/local/bin/znp`（可用 `make build` 后下发）。
2. 将生产配置放到 `/etc/znp/znp.yaml`（可由 `etc/znp-api.yaml` 复制并修改）。
3. 复制服务单元：`sudo cp deploy/systemd/znp.service /etc/systemd/system/znp.service`，按需调整 `User/Group/ZNP_CONFIG`。

## 启动与自启
```bash
sudo systemctl daemon-reload
sudo systemctl enable --now znp.service
```

## 健康检查与日志
- 健康检查：`curl -f http://127.0.0.1:8888/api/v1/ping`
- 日志：`journalctl -u znp.service -f`

> 说明：`ExecStartPre` 会运行 `znp tools check-config`，启动失败时查看 journal 日志定位配置问题。
