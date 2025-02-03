# hap-parser

Implementation of the [HAP Rule Specification](https://github.com/kyma-project/kyma-environment-broker/blob/06542126d7a9db9eec488092f19beee3b7dacf57/docs/contributor/03-11-hap-rules.md)

# Features
- Provide the rules in a YAML file :white_check_mark:
- Rule parsing and validation :white_check_mark:
- See what search labels will be used for every rule :white_check_mark:
- Rules convertion summary :white_check_mark:
- Uniqnuess & Priority :construction_worker:
- Provide example request and see which rule will be assigned :construction_worker:


# HowTo

1) Input rules:
```
cat resources/rules.yaml
 rule: 
  - aws                             # pool: hyperscalerType: aws
  - aws(PR=cf-eu11) -> EU           # pool: hyperscalerType: aws_cf-eu11; euAccess: true 
  - azure                           # pool: hyperscalerType: azure
  - azure(PR=cf-ch20) -> EU         # pool: hyperscalerType: azure_cf-ch20; euAccess: true 
  - gcp                             # pool: hyperscalerType: gcp
  - gcp(PR=cf-sa30)                 # pool: hyperscalerType: gcp_cf-sa30
  - trial -> S                      # pool: hyperscalerType: azure; shared: true - TRIAL POOL
  - sap-converged-cloud(HR=*) -> S  # pool: hyperscalerType: openstack_<HYPERSCALER_REGION>; shared: true
  - azure_lite                      # pool: hyperscalerType: azure
  - preview                         # pool: hyperscalerType: aws
  - free                            # pool: hyperscalerType: aws%        
```

2) Run rules check:
```
goo run cmd/parser/main.go
Rule No. 1 is OK, value: [aws]   - searchLabels: hyperscalerType: aws
Rule No. 2 is OK, value: [aws(PR=cf-eu11) -> EU] - searchLabels: hyperscalerType: aws_cf-eu11; euAccess: true
Rule No. 3 is OK, value: [azure] - searchLabels: hyperscalerType: azure
Rule No. 4 is OK, value: [azure(PR=cf-ch20) -> EU]       - searchLabels: hyperscalerType: azure_cf-ch20; euAccess: true
Rule No. 5 is OK, value: [gcp]   - searchLabels: hyperscalerType: gcp
Rule No. 6 is OK, value: [gcp(PR=cf-sa30)]       - searchLabels: hyperscalerType: gcp_cf-sa30
Rule No. 7 is OK, value: [trial -> S]    - searchLabels: hyperscalerType: aws; shared: true
Rule No. 8 is OK, value: [sap-converged-cloud(HR=*) -> S]        - searchLabels: hyperscalerType: openstack_*; shared: true
Rule No. 9 is OK, value: [azure_lite]    - searchLabels: hyperscalerType: azure_lite
Rule No. 10 is OK, value: [preview]      - searchLabels: hyperscalerType: aws
Rule No. 11 is OK, value: [free] - searchLabels: hyperscalerType: aws/azure
```

3) Modify input rules and check results:
```
 rule: 
  - aws                             
  - aws(PR=cf-eu11, PR=cf-eu12) -> EU           
  - azure    
```
```
Rule No. 1 is OK, value: [aws]   - searchLabels: hyperscalerType: aws
Rule No. 2 is Invalid, value: [aws(PR=cf-eu11, PR=cf-eu12) -> EU]        - reason: Input Attributes could be specified once. Attribute No. 2 is duplicated, value: PR

Rul No. 3 is OK, value: [azure] - searchLabels: hyperscalerType: azure
```