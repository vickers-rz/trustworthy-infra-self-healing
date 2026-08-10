# AIOps 告警关联、历史故障匹配与可信自动修复技术调研

> Research date: 2026-08-10  
> Project: `trustworthy-infra-self-healing`  
> Scope: Commercial AIOps / AI SRE products, OSS/open-core projects, event correlation, historical incident matching, RCA, runbook automation, policy-controlled remediation, and independent verification.

## 1. Executive summary

本报告调研的核心问题是：

> 当监控系统产生大量告警后，能否自动把告警关联为 Incident，结合拓扑、历史 Incident、Known Error、知识库和 Runbook 判断可能根因，并自动推荐或执行已有 workaround/remediation？

结论是：**可以，而且其中相当一部分能力早在 LLM 流行之前已经成熟。**

2021 年大型数据中心已经能够部署如下经典 AIOps 链路：

```text
Monitoring / Event Sources
        ↓
Normalization / Deduplication
        ↓
Event Correlation
        ↓
Topology / CMDB context
        ↓
Probable Cause / Incident
        ↓
Historical incident / KEDB / Runbook
        ↓
Operator-assisted or automated remediation
```

LLM 和现代 Agentic AI 并没有重新发明这条链路，而是显著增强了以下部分：

1. 从非结构化历史工单、聊天记录、知识文章和 Runbook 中提取操作知识；
2. 跨不同术语和文本表述寻找类似故障；
3. 在多数据源之间进行迭代调查与假设生成；
4. 生成自然语言 RCA 和处置建议；
5. 把既有 Runbook 转化为可交互的 Agent workflow。

但商业产品和开源项目普遍仍存在一个共同缺口：

> **“Agent 为什么有权执行这个动作”与“Agent 为什么认为这个动作正确”经常没有被严格拆成两个独立系统。**

这正是本项目应继续坚持的核心差异：

> **Reason probabilistically. Authorize deterministically.**

本项目不应再造完整的监控、告警管理、Runbook 编排或通用 Agent 平台，而应重点实现现有生态仍缺少的可信控制层：

```text
Evidence provenance
        ↓
Hypothesis + contradictory evidence
        ↓
Workaround applicability proof
        ↓
Deterministic risk classification
        ↓
Policy authorization
        ↓
Typed semantic action
        ↓
Trusted executor
        ↓
Independent verification
        ↓
Auditable outcome / learning
```

---

## 2. Research questions

本次研究重点回答以下问题。

### 2.1 商业产品

- 哪些商业 AIOps/ITOM/Incident Management 产品已经支持告警聚合与关联？
- 哪些产品能够自动查找类似历史 Incident？
- 哪些产品能够利用历史 resolution/workaround/runbook？
- 哪些产品支持自动执行 remediation？
- 哪些产品已经引入 LLM/Agent？
- 它们如何处理审批、权限、风险和执行边界？

### 2.2 开源和可复用项目

- 哪些能力已有成熟 OSS，不值得重新实现？
- 哪些项目只是 open-core，而关键 AI 能力属于 Enterprise？
- 哪些项目适合作为 Investigation Plane、Incident Plane、Execution Plane 或 Policy Plane？
- 哪些模块值得源码级参考，但不应该 fork？

### 2.3 对 `trustworthy-infra-self-healing` 的意义

- 我们真正的技术差异应该放在哪里？
- 哪些能力应该直接集成？
- 哪些能力应该保持自己的领域模型与安全边界？

---

## 3. Capability model

为了避免所有厂商都用“AIOps”“AI Agent”“Autonomous SRE”等营销术语，本报告采用统一能力模型。

```text
L0  Telemetry ingestion
L1  Alert normalization / deduplication
L2  Alert grouping / event correlation
L3  Topology / dependency / change correlation
L4  Historical incident similarity
L5  Probable cause / root-cause analysis
L6  Knowledge / workaround / runbook retrieval
L7  Remediation recommendation
L8  Human-approved execution
L9  Automatic execution
L10 Independent post-action verification
L11 Outcome learning / confidence calibration
```

需要特别区分：

- **Correlation**：多个告警是否属于同一个 Incident；
- **Similarity**：当前 Incident 是否与过去某个 Incident 相似；
- **RCA**：什么是当前故障的根因；
- **Workaround retrieval**：过去类似故障使用过什么处理办法；
- **Applicability**：过去的 workaround 是否真的适用于当前环境；
- **Authorization**：即使 workaround 适用，系统是否有权执行；
- **Verification**：动作执行后，是否由独立观测证明故障恢复。

商业产品通常把前五个问题整合在一个 UI 中，但对于可信自治系统，这些概念必须保持分离。

---

# Part I — Commercial products

## 4. ServiceNow ITOM + ITSM + Now Assist

### 4.1 为什么它是最重要的企业参考架构

ServiceNow 的优势不是某个单独 AI 模型，而是 ITIL 数据模型天然形成知识闭环：

```text
Alert
  ↓
Incident
  ↓
Problem
  ↓
Known Error
  ↓
Workaround
  ↓
Knowledge / Change / CMDB
```

ServiceNow 对 Known Error 的定义本身就包含已知根因和 workaround。Problem Management 用来管理 Problem 生命周期、关联 Incident，并保存相应 workaround 和 resolution。

当前 Incident Assist agentic workflow 可以查询：

- Incident 本身；
- SLA、CI、Change、Outage、Problem；
- caller 的近期 Incident；
- **similar resolved incidents**；
- on-call experts。

类似已解决 Incident 使用 semantic search 获取。这意味着“当前 Incident → 历史已解决 Incident”已经正式进入 ServiceNow 的 Agentic workflow。

### 4.2 对本项目的启示

ServiceNow 最值得借鉴的不是 LLM，而是对象边界：

```text
Incident != Problem != Known Error != Workaround != Change
```

本项目也不应该把“历史 Incident 文本”和“可执行 Runbook”混成同一种 RAG 文档。

历史 Incident 应是 **reference evidence**；Runbook 是 **candidate operational procedure**；执行权限必须来自单独的 policy/authority plane。

### 4.3 注意当前产品演进

ServiceNow 旧的 Incident Assist skill 已进入 deprecated/archive；新的 **Incident Assist agentic workflow** 接管并增强了类似能力。因此研究时应以 agentic workflow 为当前架构，而不是旧 skill。

### Primary sources

- [Incident Assist agentic workflow](https://www.servicenow.com/docs/r/it-service-management/now-assist-for-it-service-management-itsm/now-assist-itsm-incident-assist-workflow.html)
- [Problem Management overview](https://www.servicenow.com/docs/en-US/bundle/zurich-it-service-management/page/product/problem-management/concept/c_ProblemManagement.html)
- [What is Problem Management / Known Error / Workaround](https://www.servicenow.com/products/itsm/what-is-problem-management.html)

---

## 5. BigPanda

BigPanda 是本次调研中与“**自动关联已有 workaround / 历史处置经验**”最接近的商业产品之一。

### 5.1 Similar Incidents

BigPanda Similar Incidents 会按照以下类别计算相似度：

- Entity；
- Problem；
- Impact；
- Topology。

官方文档明确显示历史 Incident 可以展示：

- impact；
- assignment；
- resolution steps；
- root cause；
- 为什么被认为相似；
- similarity score。

这比单纯 embedding 相似度更值得注意，因为它已经在做某种**结构化 applicability approximation**：不仅比较文本，还比较实体、故障类型、影响和拓扑。

### 5.2 AI Incident Analysis

Advanced Insight Module 使用 LLM 生成：

- incident title；
- summary；
- probable root cause；
- recommended actions。

### 5.3 L1 Agent

2026 年 BigPanda L1 Agent 已明确采用 Agentic L1 Operations 方向。其文档描述了两类 Runbook Automation：

1. **Context collection**：读取既有 Runbook，将其拆成可执行步骤，执行诊断并把结果反馈给 reasoning；
2. **Remediation actions**：执行 restart service、clear queue、known fix 等修复动作。

其中 remediation 明确需要 approval workflow。

需要避免过度解读：官方页面同时把部分 suppression、prioritization、escalation、auto-resolution 列为 upcoming capabilities，因此不能把所有路线图能力都视为已经 GA。

### 5.4 对本项目的启示

BigPanda 给出的最大启示是：

> **历史 Incident similarity 本身已经不是创新点。**

真正值得继续推进的是：

```text
similarity
   ↓
applicability proof
   ↓
current-state evidence
   ↓
risk / authority
```

即：相似并不等于可以执行同一个 workaround。

### Primary sources

- [Similar Incidents](https://docs.bigpanda.io/en/similar-incidents)
- [L1 Agent](https://docs.bigpanda.io/en/l1-agent)
- [Advanced Insight Module](https://docs.bigpanda.io/en/advanced-insight-module)

---

## 6. PagerDuty AIOps + Runbook Automation

PagerDuty 已经形成一条相当完整的 Incident → historical context → automation 链。

### 6.1 Past Incidents

Past Incidents 使用机器学习在**同一个 service**的历史 Incident 中寻找相似项，考虑：

- Incident title semantics；
- responders；
- incident duration；
- creation date/time。

并显示过去采取过的 remediation steps。

这一约束非常重要：PagerDuty Past Incidents 并不是任意跨组织语义搜索，而是利用 Service 边界缩小候选空间。

### 6.2 Related Incidents

Related Incidents 面向当前正在发生的其他 Incident，结合实时机器学习和 service dependency data，回答：

> 当前其他服务上还有哪些 Incident 可能与本 Incident 相关？

它与 Past Incidents 分别解决“历史相似性”和“当前横向影响”。

### 6.3 Probable Origin

Probable Origin 使用历史 Incident occurrence pattern 估计某个 Incident 是否可能引发另一个 Incident，补充 Related Incidents 的实时关联视角。

### 6.4 Event Orchestration + Automation Actions

PagerDuty Event Orchestration 可以：

- 对事件进行 enrichment、routing、grouping；
- 触发 webhook；
- 触发已有 Automation Action；
- 在 human notification 之前自动运行 diagnostics/remediation。

官方给出的闭环示例是：

```text
Alert arrives
   ↓
Pause notification
   ↓
Run remediation
   ↓
Alert resolves?
   ├─ yes → do not page human
   └─ no  → trigger incident normally
```

这是传统 closed-loop remediation 的成熟实现。

### 6.5 对本项目的启示

PagerDuty 再次证明：

- Runbook catalog；
- incident-triggered automation；
- pre-human remediation；

都不应成为本项目重新实现的重点。

我们的差异应放在 remediation action 被触发之前的**证据、适用性、风险和确定性授权**。

### Primary sources

- [Past Incidents](https://support.pagerduty.com/main/docs/past-incidents)
- [Related Incidents](https://support.pagerduty.com/main/docs/related-incidents)
- [Event Orchestration](https://support.pagerduty.com/main/docs/event-orchestration)
- [Event Orchestration examples](https://support.pagerduty.com/main/docs/event-orchestration-examples)
- [Automation Actions](https://support.pagerduty.com/main/docs/automation-actions)

---

## 7. Dynatrace Intelligence

Dynatrace 最值得研究的是它没有把 RCA 完全建立在语言模型或时间相关性上。

### 7.1 Causal/topology-oriented RCA

Dynatrace Intelligence 当前明确采用 deterministic, causation-based analysis，并使用 Smartscape / topology / dependency context 进行 RCA。

其事件处理过程包含：

```text
Ingestion
  ↓
Normalization
  ↓
Topology creation
  ↓
Aggregation
  ↓
Deduplication
  ↓
Causal RCA / impact analysis
```

官方 RCA 文档特别强调：**仅依靠 time correlation 不足以确定 root cause**。系统会沿服务、进程、主机、应用和事务依赖进行分析，并排名 root cause contributors。

### 7.2 Blast radius

Dynatrace 还对受影响的 application/service/SLO/user 做 impact analysis，明确把 blast radius 作为 Incident triage 的一部分。

这与本项目未来 deterministic risk classifier 中的 blast radius 很接近。

### 7.3 GenAI/Agentic layer

GenAI 层用于：

- problem summary；
- root-cause explanation；
- suggested remediation steps；
- troubleshooting guide discovery。

当前更高级的 approved agentic actions / rollback 等部分官方仍标注 Preview，因此应把“成熟 causal RCA”和“仍快速演进的 Agentic execution”分开评价。

### 7.4 对本项目的启示

推荐采用类似思想：

```text
Topology / deterministic facts
        +
Statistical signals
        +
LLM semantic reasoning
        ↓
Evidence-grounded hypothesis
```

而不是：

```text
all logs → LLM → root cause
```

### Primary sources

- [Dynatrace Intelligence](https://docs.dynatrace.com/docs/dynatrace-intelligence)
- [Root cause analysis concepts](https://docs.dynatrace.com/docs/dynatrace-intelligence/root-cause-analysis/concepts)
- [Event analysis and correlation](https://docs.dynatrace.com/docs/dynatrace-intelligence/root-cause-analysis/event-analysis-and-correlation)
- [Agentic and generative AI](https://docs.dynatrace.com/docs/dynatrace-intelligence/agentic-and-generative-ai)

---

## 8. IBM Netcool Operations Insight

IBM Netcool 代表了传统大型数据中心 AIOps 的经典路线：

```text
Events
  ↓
Event analytics
  ↓
Grouping / topology association
  ↓
Probable cause scoring
  ↓
Incident / operator workflow
  ↓
Runbook automation
```

IBM 的 probable-cause analysis 可针对 topologically grouped events 自动进行，并通过权重和 criteria 对候选 cause event 评分。

这种设计说明，在 LLM 出现以前，AIOps 已经能使用：

- topology；
- event grouping；
- rule/weighting；
- probable-cause scoring；
- runbook automation。

因此，2021 年在大型银行数据中心监控大厅看到“告警自动关联、推荐历史处理经验或 Runbook”在技术上完全合理，它不需要依赖今天的 LLM Agent 才能实现。

### Primary source

- [IBM Netcool Operations Insight — customizing probable cause](https://www.ibm.com/docs/en/noi/1.6.13?topic=cause-customizing-probable)

---

## 9. AWS CloudWatch Investigations

CloudWatch Investigations 是现代“AI investigation → hypothesis → runbook”路线中非常值得研究的实现。

### 9.1 Investigation data

它可以扫描并关联：

- metrics；
- logs；
- deployment events；
- CloudTrail change events；
- X-Ray traces；
- AWS Health events；
- topology-related information。

并生成 observation、suggestion 和 root-cause hypothesis。

### 9.2 Hypothesis is not execution

CloudWatch 提供 Accept / Discard hypothesis，接受 hypothesis 并不会自动执行 runbook。

当存在 suggested action 时，用户还要：

1. review reasoning；
2. review runbook；
3. review parameters；
4. preview impact；
5. explicitly execute。

这形成了一条比“LLM 给出 shell 命令”安全得多的路径。

### 9.3 Existing runbook selection

CloudWatch 可以建议 AWS-owned 或 customer-owned Systems Manager Automation runbooks。文档说明自定义 runbook 候选会通过 runbook keyword 与 Incident 相关 terms 比较生成。

这说明即使在商业产品里，“Runbook applicability”也仍可以相对朴素，并未完全解决当前环境适配证明的问题。

### 9.4 Audit and identity

Investigation 使用登录用户权限，动作通过 IAM 授权；investigation action 记录在 CloudTrail。这一点与本项目 `ExecutionContext` / approval provenance / audit boundary 思路高度一致。

### Primary sources

- [CloudWatch investigations](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/Investigations.html)
- [Suggested runbook remediations](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/suggested-investigation-actions.html)

---

## 10. Azure SRE Agent

Azure SRE Agent 是当前商业产品中与本项目安全控制方向最接近的系统之一。

### 10.1 Investigation and mitigation

Microsoft 将其定义为用于 diagnosis、root cause investigation、mitigation proposal 和 runbook-driven response 的可靠性 Agent。

### 10.2 Review vs Autonomous

当前有两类 run mode：

- **Review**：Agent 提出动作，用户 approve/deny；
- **Autonomous**：Agent 自动执行并报告结果。

关键是 permissions 与 run mode 分离：

```text
Permission → can this identity access the resource?
Run mode   → must a human approve this action?
```

两者都满足时动作才能发生。

### 10.3 Tool access policies

Azure 还增加了独立 tool access policy，可表达：

- allow；
- ask；
- deny。

全局 policy 可以禁止工具，即使 Agent 自己或 thread 想扩大权限也不能弱化全局 deny。

### 10.4 Hard safety guards

当前官方文档列出若干 command-level guardrails，例如：

- delete/remove operation blocked；
- Key Vault command blocked；
- management lock respected；
- subscription ID validation。

这再次证明“把安全判断从 LLM prompt 中移出去”已经成为成熟 Agentic Ops 产品的重要方向。

### 10.5 Verification

Microsoft 文档还明确描述 Agent 在 mitigation 后执行 verification 并报告结果。

不过本项目仍应比这一模式更严格：**verification 应由独立 verifier / watchdog 的观测逻辑完成，而不是仅由同一 reasoning agent 自我确认。**

### Primary sources

- [Azure SRE Agent documentation](https://learn.microsoft.com/en-us/azure/sre-agent/)
- [Run modes](https://learn.microsoft.com/en-us/azure/sre-agent/run-modes)
- [Execute mitigations](https://learn.microsoft.com/en-us/azure/sre-agent/execute-mitigations)
- [Tool access policies](https://learn.microsoft.com/en-us/azure/sre-agent/tool-access-policies)

---

# Part II — Open source / open-core ecosystem

## 11. HolmesGPT

**Role:** Investigation Plane / SRE Agent  
**License:** Apache-2.0  
**Status:** CNCF Sandbox project

HolmesGPT 是当前最值得直接集成或源码级学习的 OSS SRE Agent 之一。

### 11.1 Core architecture

源码核心包括：

```text
holmes/core/tool_calling_llm.py
holmes/core/tools_utils/tool_executor.py
holmes/core/prompt.py
holmes/plugins/toolsets/
```

其核心模式是多轮 tool-calling investigation：

```text
Incident / question
      ↓
LLM hypothesis
      ↓
Tool call
      ↓
Observation
      ↓
Updated hypothesis
      ↓
Next tool call
```

这非常适合承担本项目的 probabilistic Investigation Plane。

### 11.2 Kubernetes Remediation MCP

HolmesGPT 当前 Kubernetes Remediation MCP 设计已经非常值得研究。

其核心安全思想是：**通过 tool identity 区分 auto-approved 和 approval-gated action，而不是让模型根据参数猜测风险。**

设计中的工具分类包括：

| Tool | Mutating | Approval |
|---|---:|---|
| `read_file_from_container` | no | auto |
| `run_preapproved_kubectl_command` | no | auto |
| `run_preapproved_diagnostic_image` | diagnostic | auto |
| `run_kubectl_command` | yes / arbitrary | human approval |

MCP server 自己负责：

- command allowlist；
- dangerous-flag blocklist；
- path allow/deny；
- image allowlist；
- shell metacharacter rejection；
- timeout；
- scoped Kubernetes RBAC。

### 11.3 最重要的 residual risk

HolmesGPT 自己的设计文档明确指出一个关键问题：

> approval boundary 依赖 HolmesGPT 与 MCP server 配置保持一致；server 本身并不知道一个调用是否真的获得了人类 approval。

如果有其他客户端能够直接调用 remediation MCP server，就可能绕过 HolmesGPT approval layer。

**这正好验证本项目现有架构的必要性：Executor 不应仅仅信任“调用来自 Agent”。它应该要求可验证的 authorization artifact / trusted ExecutionContext，并由 deterministic guard 独立检查。**

### 11.4 Adoption decision

**不要 fork HolmesGPT。**

推荐将其作为外部 Investigation Agent：

```text
HolmesGPT
    ↓
Evidence adapter / structured output
    ↓
InfraSelfHeal Evidence + Hypothesis contracts
```

LLM/Agent 生成 hypothesis，但不能生成 authoritative evidence provenance，也不能直接获得 executor authority。

### Primary sources

- [HolmesGPT repository](https://github.com/HolmesGPT/holmesgpt)
- [Kubernetes Remediation MCP design spec](https://github.com/HolmesGPT/holmesgpt/blob/master/specs/kubernetes-remediation-mcp.md)

---

## 12. Keep

**Role:** Event / Alert / Incident Plane  
**License:** MIT for core; proprietary license under `ee/`  
**Model:** open-core

Keep 是非常适合作为 alert normalization / incident management frontend 的项目，但必须明确其 OSS 与 Enterprise 边界。

### 12.1 OSS capabilities

核心能力包括：

- multi-provider alert ingestion；
- alert feed；
- dedup/noise reduction；
- manual correlation rules；
- topology correlation；
- incidents；
- workflows；
- provider abstraction。

Manual Correlation Engine 可以基于 alert attributes 用逻辑表达式创建 Incident 或 Incident Candidate。

Topology Processor 可以根据 service/application/dependency topology 把多个相关告警聚合为 application-level Incident。

### 12.2 AI licensing boundary

Keep 官方当前明确：

```text
AI Correlation
Keep Cloud                yes
Keep Enterprise On-Prem   yes
Keep Open Source          no
```

完整 AI correlation 使用历史 alert/incident 数据训练 proprietary model，并自动将未分配 alerts 聚类到 Incident。

另一方面，Semi-Automatic AI Correlation 在 OSS 中属于 experimental：用户选中 alerts 后，LLM 生成 Incident candidates，再由人工确认。

### 12.3 Repository license boundary

仓库根 LICENSE 明确：

- `ee/` 之外内容使用 MIT；
- `ee/` 使用单独 Enterprise License；
- EE 生产使用要求有效订阅。

因此不能把“GitHub 仓库公开”理解为“全部功能可自由商用”。

### 12.4 Adoption decision

Keep 可作为：

```text
Alert / Incident UI
Normalization
Manual/topology correlation
Workflow integration
```

但本项目核心不应依赖 Keep proprietary AI correlator。

应该保持自己的 adapter：

```text
IncidentSource
  ├─ KeepAdapter
  ├─ PagerDutyAdapter
  ├─ ServiceNowAdapter
  └─ AlertmanagerAdapter
```

### Primary sources

- [Keep repository](https://github.com/keephq/keep)
- [Keep core LICENSE](https://github.com/keephq/keep/blob/main/LICENSE)
- [Keep EE LICENSE](https://github.com/keephq/keep/blob/main/ee/LICENSE)
- [Manual Correlation Rules](https://docs.keephq.dev/overview/correlation-rules)
- [Topology Correlation](https://docs.keephq.dev/overview/correlation-topology)
- [AI Correlation](https://docs.keephq.dev/overview/ai-correlation)
- [AI Semi Automatic Correlation](https://docs.keephq.dev/overview/ai-semi-automatic-correlation)

---

## 13. Robusta

**Role:** Kubernetes Evidence Collection + deterministic remediation  
**License:** MIT

Robusta 的优势是 Kubernetes-native alert enrichment 与 playbook automation。

其 Playbook 是经典：

```text
Trigger
  ↓
Action(s)
  ↓
Finding / notification / remediation
```

可以：

- 收集 Pod logs；
- 收集 Kubernetes events；
- enrich Prometheus alert；
- 自动运行 Kubernetes Job；
- 执行 Bash remediation；
- restart rollout；
- cordon/uncordon/drain node；
- delete stuck Pod/Job。

它证明 Kubernetes self-healing 的 deterministic action layer 已经相当成熟。

### Adoption decision

不把 Robusta 作为总体 control plane，而是作为：

```text
Kubernetes
   ↓
Robusta
   ↓
Evidence adapter
   ↓
EvidenceBundle
```

或者作为特定 low-risk deterministic remediation backend。

### Primary sources

- [Robusta repository](https://github.com/robusta-dev/robusta)
- [Automatic Remediation](https://docs.robusta.dev/master/playbook-reference/automatic-remediation-examples/index.html)
- [Remediation actions](https://docs.robusta.dev/docs-ui-sink-demo/playbook-reference/actions/remediation.html)

---

## 14. StackStorm

**Role:** General Event-Driven Execution Plane  
**License:** Apache-2.0

StackStorm 是最不值得重新实现的一类基础设施。

经典模型：

```text
Sensor
  ↓
Trigger
  ↓
Rule
  ↓
Action / Workflow
```

它已经处理：

- external event ingestion；
- trigger/rule matching；
- workflow orchestration；
- action execution；
- integrations；
- execution history；
- auto-remediation。

### Adoption decision

对于 Kubernetes 之外的 Linux/VM/Cloud/Network/OpenStack/VMware 等系统，StackStorm 很适合做外部 execution backend。

但 InfraSelfHeal 不应该允许 LLM 给 StackStorm 任意 command。

正确接口应是 typed semantic action：

```json
{
  "action": "restart_workload",
  "target": "payment-api",
  "scope": "prod-eu-1"
}
```

由受信任 adapter 映射到：

```text
StackStorm action/workflow ID + validated parameters
```

### Primary source

- [StackStorm/st2](https://github.com/StackStorm/st2)

---

## 15. Rundeck Community

**Role:** Runbook Catalog / Human-centric Operations / Execution  
**License:** Apache-2.0

Rundeck 的自然抽象是：

```text
Project
  ↓
Job
  ↓
Workflow
  ↓
Steps
  ↓
Nodes
```

它非常适合：

- 运维 Runbook catalog；
- operator self-service；
- push-button diagnostics/remediation；
- node execution；
- execution history；
- workflow steps 和 plugins。

### StackStorm vs Rundeck

| Need | Better natural fit |
|---|---|
| Event-driven automation | StackStorm |
| Automatic remediation | StackStorm |
| Human-run SOP | Rundeck |
| Runbook catalog | Rundeck |
| Operator self-service | Rundeck |
| Node-oriented job execution | Rundeck |

MVP 不需要同时集成两者。

### Primary source

- [rundeck/rundeck](https://github.com/rundeck/rundeck)

---

## 16. PyRCA

**Role:** Statistical / causal RCA evidence generator  
**License:** BSD-3-Clause

PyRCA 不是 Incident platform，也不是 Agent。它是 Root Cause Analysis library。

它提供：

- anomaly detection pipeline；
- causal graph discovery；
- RCA algorithms；
- root-cause ranking；
- domain knowledge integration。

源码扩展接口明确要求 RCA analyzer 实现 `train()` 与 `find_root_causes()`，并返回结构化 `RCAResults`。

### Adoption decision

PyRCA 不应该决定执行动作，而应该产生独立证据信号：

```text
PyRCA candidate root causes
           ↓
EvidenceBundle / Diagnosis input
           ↑
HolmesGPT hypothesis
           ↑
Topology / change evidence
```

它的最大价值是让 LLM hypothesis 有一个**独立的 statistical/causal comparator**。

### Primary source

- [salesforce/PyRCA](https://github.com/salesforce/PyRCA)

---

## 17. Open Policy Agent (OPA)

**Role:** General-purpose external policy engine  
**License:** Apache-2.0  
**Status:** CNCF Graduated

OPA 可以作为未来 policy backend 的候选，但本项目不必立即把所有 deterministic guard 逻辑迁移到 Rego。

OPA 的价值是把 authorization policy 与 application code 分离，并支持统一 context-aware policy enforcement。

### Adoption decision

短期：

- 继续维护项目当前 Go deterministic Guard；
- 保持 Policy Engine interface。

中期：

- 评估 OPA/Rego backend；
- 将组织级 policy、environment policy、change window、service criticality 等外部规则迁移为 GitOps-managed policy bundle。

关键原则仍然是：

> OPA 可以决定 authority，但不能把 LLM confidence 当作绕过 deny 的权限来源。

### Primary source

- [Open Policy Agent](https://github.com/open-policy-agent/opa)

---

# Part III — Cross-product findings

## 18. What is already mature

以下能力已经有成熟商业实现或成熟 OSS，原则上不应成为本项目的主要研发投入。

### 18.1 Do not rebuild

- alert ingestion；
- provider connectors；
- alert normalization；
- basic deduplication；
- rule-based grouping；
- topology-based grouping；
- generic workflow engine；
- generic runbook execution；
- SSH / Kubernetes / cloud command execution；
- RBAC system；
- notification / on-call system；
- generic Incident UI；
- raw LLM tool-calling loop。

### 18.2 Mature but worth integrating/evaluating

- historical Incident similarity；
- change correlation；
- topology-aware RCA；
- statistical RCA；
- LLM investigation；
- Runbook retrieval；
- approval workflow。

---

## 19. What is still insufficiently solved

### 19.1 Similarity is not applicability

现有系统常见逻辑：

```text
current incident
    ↓
similar historical incident
    ↓
previous resolution
    ↓
recommended action
```

问题是：

```text
same symptom ≠ same cause
same cause   ≠ same environment
same runbook ≠ currently safe
```

因此本项目应增加明确的 **Workaround Applicability** 概念。

候选条件包括：

```text
service identity
version / deployment revision
runtime/environment
resource type
statefulness
topology/dependency shape
configuration assumptions
current symptom
current root-cause hypothesis
recent change history
blast radius
business criticality
runbook version
historical success/failure
```

输出不能只有 similarity score，而应该包含可审计 applicability explanation。

---

## 20. Evidence provenance is a product gap

商业 Agent 往往会展示 reasoning，但“展示 reasoning”并不等于**具备可验证的 evidence authority**。

本项目当前 `EvidenceBundle` 设计已经正确地把以下字段从 planner 输出中剥离：

- source；
- collector；
- observation time；
- collection time；
- freshness；
- trust metadata；
- integrity digest。

LLM 只能引用 evidence ID，不能自己创建“可信来源”。

这是相比普通 RAG/Agent 平台更强的安全边界。

---

## 21. Approval is not authorization

HolmesGPT Kubernetes Remediation MCP 展示了一个非常典型的问题：

```text
Agent UI asks human for approval
        ↓
MCP server receives command
```

如果 MCP server 自己不能验证 approval provenance，则“approval”仍然可能只是上游 UI 协议，而不是 executor 的强安全边界。

本项目应保持：

```text
Proposal
+
ExecutionContext
+
EvidenceBundle
        ↓
Deterministic Guard
        ↓
Authorization artifact
        ↓
Executor
```

未来 mutation-capable executor 最好要求不可伪造/可审计绑定的 authorization artifact，而不是只信任网络调用方。

---

## 22. LLM should select semantic capabilities, not commands

错误接口：

```text
LLM → "kubectl delete ..."
LLM → "rm -rf ..."
LLM → arbitrary shell
```

推荐接口：

```yaml
action:
  type: restart_workload
  target:
    kind: Deployment
    namespace: payments
    name: payment-api
```

trusted executor 再映射到 Kubernetes API、StackStorm、Rundeck、SSM Automation 等具体 backend。

这一点可以：

- 限制 action space；
- 做 deterministic validation；
- 做 blast-radius classification；
- 做 backend-independent policy；
- 生成稳定 audit record；
- 支持 replay 和 simulation。

---

## 23. Verification should be independent

很多产品已经能做到“执行后再检查”，但可信系统应该进一步区分：

```text
Planner / Investigator
        ↓
Executor
        ↓
Independent Verifier
```

Verifier 应使用外部可观察 postconditions：

- original alert state；
- SLO；
- error rate；
- latency；
- saturation；
- readiness；
- dependency health；
- synthetic transaction；
- application invariant。

不应该允许 proposing model 用自己的语言输出宣告“已经修复”。

---

# Part IV — Recommended project positioning

## 24. Proposed system role

本项目不应定位成：

> Another AIOps platform

也不应定位成：

> Another SRE chatbot

更准确的定位是：

> **Trust and Safety Control Plane for Agentic Operations**

或：

> **Evidence-Grounded, Policy-Controlled Autonomous Remediation Framework**

核心价值不是“AI 能修服务器”，而是：

> **系统可以证明为什么这个动作在当前证据、风险、权限和策略下被允许执行，并能独立证明执行后的结果。**

---

## 25. Recommended integration architecture

```text
Prometheus / OTel / Zabbix / Cloud / ITSM
                  │
                  ▼
       Keep / Alertmanager / Adapter
                  │
            Incident/Event
                  │
                  ▼
            Evidence Plane
        ┌─────────┼──────────┐
        │         │          │
      Robusta   topology   changes/history
        │         │          │
        └─────────┼──────────┘
                  ▼
            EvidenceBundle
                  │
          ┌───────┴────────┐
          ▼                ▼
     HolmesGPT           PyRCA
 semantic investigation statistical/causal
          │                │
          └───────┬────────┘
                  ▼
         Hypothesis candidates
                  │
                  ▼
       Workaround Applicability
                  │
                  ▼
       Deterministic Risk Classifier
                  │
                  ▼
       Guard / OPA-compatible Policy
                  │
          ┌───────┼──────────┐
          ▼       ▼          ▼
        DENY    REVIEW      AUTO
                  │          │
                  └────┬─────┘
                       ▼
              Typed semantic action
                       │
             ┌─────────┼─────────┐
             ▼         ▼         ▼
        Kubernetes  StackStorm  SSM/Rundeck
                       │
                       ▼
                 Infrastructure
                       │
                       ▼
              Independent Verifier
                       │
              success / rollback
                       │
                       ▼
                 Outcome record
```

---

## 26. Reuse matrix

| Component | Best candidate | Recommendation |
|---|---|---|
| Alert ingestion/UI | Keep / existing monitoring | Integrate, do not rebuild |
| Incident source | Keep / PagerDuty / ServiceNow adapter | Keep interface generic |
| K8s evidence enrichment | Robusta | Optional adapter |
| LLM investigation | HolmesGPT | Integrate, do not fork initially |
| Statistical RCA | PyRCA | Add in Phase 2 |
| Policy engine | Current Go Guard, later OPA option | Keep authority deterministic |
| Generic execution | StackStorm | Preferred first generic backend |
| Human runbook catalog | Rundeck | Optional later backend |
| Kubernetes execution | typed Kubernetes adapter | Keep narrow semantic actions |
| Cloud execution | SSM / provider-native automation | Adapter |
| Verification | **InfraSelfHeal** | Core project responsibility |
| Evidence provenance | **InfraSelfHeal** | Core project responsibility |
| Workaround applicability | **InfraSelfHeal** | Core project responsibility |
| Risk classification | **InfraSelfHeal** | Core project responsibility |
| Authorization artifact | **InfraSelfHeal** | Core project responsibility |

---

## 27. License / product boundary summary

| Project | License/model | Important boundary |
|---|---|---|
| HolmesGPT | Apache-2.0 | OSS; remediation security still depends on integration boundary |
| Robusta | MIT | OSS |
| StackStorm | Apache-2.0 | OSS |
| Rundeck Community | Apache-2.0 | OSS community core |
| PyRCA | BSD-3-Clause | OSS library |
| OPA | Apache-2.0 | OSS, CNCF Graduated |
| Keep core | MIT | `ee/` excluded |
| Keep EE | Proprietary Enterprise License | production use requires subscription |
| Keep full AI correlation | Cloud / Enterprise | not available in OSS |
| Keep semi-auto AI correlation | OSS experimental | human-review oriented |

License classification must be checked again before redistribution or embedding dependencies into a commercial product. Public source availability alone does not imply an OSI-compatible open-source license.

---

# Part V — MVP implications

## 28. Recommended Phase 1

第一阶段不要同时集成所有项目。

建议链路：

```text
Prometheus / Kubernetes
        ↓
Alertmanager or Keep
        ↓
Evidence collector
        ↓
HolmesGPT investigation
        ↓
InfraSelfHeal EvidenceBundle + Hypothesis
        ↓
Workaround candidate
        ↓
Current deterministic Guard
        ↓
Typed low-risk execution
        ↓
Independent verification
```

Phase 1 的目标不是“全自动 SRE”，而是完整证明：

```text
alert
→ evidence
→ hypothesis
→ historical workaround
→ applicability
→ authorization
→ execution
→ verification
→ audit record
```

---

## 29. Recommended Phase 2

加入：

- PyRCA；
- topology graph；
- change correlation；
- multiple hypothesis ranking；
- contradicting evidence；
- confidence calibration。

形成：

```text
LLM semantic RCA
vs
statistical RCA
vs
topology RCA
vs
historical incident similarity
```

并记录各类 signal 与最终 outcome 的 calibration。

---

## 30. Recommended Phase 3

开始有限 autonomous remediation。

建议 action policy 至少分三类：

### AUTO_SAFE

示例：

- restart a single unhealthy stateless replica；
- recreate disposable diagnostic workload；
- clear explicitly disposable cache；
- restart failed sidecar；
- scale within a pre-authorized narrow range。

### REQUIRE_APPROVAL

示例：

- restart application deployment；
- scale beyond routine threshold；
- restart connection pool；
- fail over non-core replica；
- change production routing weight。

### NEVER_AUTO / hard deny

示例：

- destructive storage operation；
- persistent-volume deletion；
- database schema destructive change；
- credential/key mutation；
- IAM privilege expansion；
- arbitrary firewall mass change；
- uncontrolled core database failover。

最终类别必须由 deterministic policy 和可信上下文决定，而不是 LLM 自报 risk level。

---

# 31. Final conclusions

### Conclusion 1 — 2021-era AIOps already solved much of the basic workflow

告警降噪、事件关联、拓扑相关性、probable cause、KEDB/Runbook 和 closed-loop automation 并不是 LLM 时代才出现。

### Conclusion 2 — LLM makes unstructured operational memory usable

现代 LLM/Agent 的主要增量是把工单、Runbook、知识库、聊天、代码、Change 等非结构化信息纳入实时调查，并进行多轮工具调用。

### Conclusion 3 — “find a similar incident and reuse its workaround” is no longer enough as a project differentiator

ServiceNow、BigPanda、PagerDuty 等已经具备成熟或接近成熟的历史 Incident 检索/相似性能力。

### Conclusion 4 — execution orchestration is commodity infrastructure

StackStorm、Rundeck、Robusta、SSM Automation、PagerDuty Automation 等已经提供成熟执行层。本项目不应重新实现通用 Runbook engine。

### Conclusion 5 — trustworthy authorization is still the strongest project opportunity

真正值得继续推进的是：

```text
trusted evidence provenance
+
explicit uncertainty
+
contradictory evidence
+
workaround applicability
+
deterministic risk
+
external policy authority
+
typed semantic action
+
independent verification
```

最终目标应始终保持：

> **The model may form beliefs. It must not manufacture authority.**

> **Reason probabilistically. Authorize deterministically.**

---

# 32. References

## Commercial / cloud products

1. ServiceNow — Incident Assist agentic workflow: https://www.servicenow.com/docs/r/it-service-management/now-assist-for-it-service-management-itsm/now-assist-itsm-incident-assist-workflow.html
2. ServiceNow — Problem Management: https://www.servicenow.com/docs/en-US/bundle/zurich-it-service-management/page/product/problem-management/concept/c_ProblemManagement.html
3. BigPanda — Similar Incidents: https://docs.bigpanda.io/en/similar-incidents
4. BigPanda — L1 Agent: https://docs.bigpanda.io/en/l1-agent
5. PagerDuty — Past Incidents: https://support.pagerduty.com/main/docs/past-incidents
6. PagerDuty — Related Incidents: https://support.pagerduty.com/main/docs/related-incidents
7. PagerDuty — Event Orchestration: https://support.pagerduty.com/main/docs/event-orchestration
8. Dynatrace — Dynatrace Intelligence: https://docs.dynatrace.com/docs/dynatrace-intelligence
9. Dynatrace — RCA concepts: https://docs.dynatrace.com/docs/dynatrace-intelligence/root-cause-analysis/concepts
10. IBM Netcool — probable cause: https://www.ibm.com/docs/en/noi/1.6.13?topic=cause-customizing-probable
11. AWS — CloudWatch investigations: https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/Investigations.html
12. AWS — suggested runbook remediation: https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/suggested-investigation-actions.html
13. Microsoft — Azure SRE Agent: https://learn.microsoft.com/en-us/azure/sre-agent/
14. Microsoft — Azure SRE Agent run modes: https://learn.microsoft.com/en-us/azure/sre-agent/run-modes
15. Microsoft — execute mitigations: https://learn.microsoft.com/en-us/azure/sre-agent/execute-mitigations

## Open source / open-core

16. HolmesGPT: https://github.com/HolmesGPT/holmesgpt
17. HolmesGPT Kubernetes Remediation MCP spec: https://github.com/HolmesGPT/holmesgpt/blob/master/specs/kubernetes-remediation-mcp.md
18. Keep: https://github.com/keephq/keep
19. Keep AI Correlation: https://docs.keephq.dev/overview/ai-correlation
20. Keep Manual Correlation: https://docs.keephq.dev/overview/correlation-rules
21. Robusta: https://github.com/robusta-dev/robusta
22. Robusta Automatic Remediation: https://docs.robusta.dev/master/playbook-reference/automatic-remediation-examples/index.html
23. StackStorm: https://github.com/StackStorm/st2
24. Rundeck: https://github.com/rundeck/rundeck
25. PyRCA: https://github.com/salesforce/PyRCA
26. Open Policy Agent: https://github.com/open-policy-agent/opa

---

## 33. Maintenance note

该领域 2025–2026 年更新非常快，尤其是 Agentic remediation、tool approval、MCP integration 和 autonomous execution。后续更新本报告时，应优先核查：

- feature 是否 GA / Preview / Limited Availability；
- commercial feature 是否进入 OSS；
- license 是否发生变化；
- Agent write-action 是否增加新的 server-side guard；
- approval 是否由 executor 强验证；
- 是否出现真正独立的 verification / authorization artifact；
- historical incident matching 是否增加环境/version/topology applicability validation。
