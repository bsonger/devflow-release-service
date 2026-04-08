# Release 类型速查

- id: string
- execution_intent_id: string
- application_id: string
- application_name: string
- image_id: string
- image_name: string
- project_name: string
- env: string
- type: string
- status: Pending | Running | Succeeded | Failed | RollingBack | RolledBack | Syncing | SyncFailed
- steps: []ReleaseStep

## ReleaseStep
- name: string
- progress: int
- status: Pending | Running | Succeeded | Failed
- message: string
- start_time: string
- end_time: string
