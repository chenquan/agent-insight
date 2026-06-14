## ADDED Requirements

### Requirement: calculateHotspots 除零防护
`calculateHotspots` SHALL 在 totalValue == 0 时返回空 hotspots 切片而非产生 NaN 百分比。

#### Scenario: 所有 sample 被 focus/ignore 过滤
- **WHEN** profile 有样本但所有样本被 focus/ignore 模式过滤掉
- **THEN** `calculateHotspots` 返回空切片，不报错，不产生 NaN

#### Scenario: 正常 profile 不受影响
- **WHEN** profile 包含样本且无过滤或过滤后仍有样本
- **THEN** 行为与修改前完全一致
