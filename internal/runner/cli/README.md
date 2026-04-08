# CLI 运行器（占位说明）

本目录包含用于以 **命令行（CLI）** 方式直接控制 rclone 的运行器：
- 启动 rclone 子进程（exec）
- 解析进度（--stats-one-line / --use-json-log）
- 优雅停止（SIGINT → SIGTERM → SIGKILL）
- 将进度写入内存态供 /runs/active 查询

当前已提交最小可用实现，后续将逐步接入服务层并替代 RC 路径。
