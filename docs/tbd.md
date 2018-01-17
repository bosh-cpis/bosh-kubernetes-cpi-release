# Blue-green cluster upgrade

Challenges:

- how do you connect overlay networking into a new cluster
  - kubernetes may start giving out same cluster IPs -> configure different ranges
  - all depends on underlying overlay networking technology -> too different

# CF deploy/upgrade

Challenges:

- how to resurrection ~10 things in parallel? (10 workers?)
- lowering `diego.rep.evacuation_timeout_in_seconds` on diego cells

More > 2mins (not sure why)

```
database: database/51d11a94-5cb2-4948-9390-15abb632c48f (0) (00:09:13)
uaa: uaa/e49ae1bf-ca45-4958-8d3c-ea41271bec95 (0) (00:04:40)
uaa: uaa/27decf00-79c1-463d-ae49-2ddba09b4456 (1) (00:03:16)
api: api/41cbe7a8-f3de-43fd-b12c-7c0870198667 (0) (00:07:03)
singleton-blobstore: singleton-blobstore/1ce231ba-0978-4ecb-b48c-f53e95a26dda (0) (00:03:34)
```

Less < 2mins

```
diego-cell: diego-cell/8d82ba9c-44c2-4474-a854-60bb729421a3 (0) (00:02:00)
cc-worker: cc-worker/6002fb2c-9a09-48d7-ba48-de663bea7105 (0) (00:01:28)
api: api/da223536-fa4e-4c0a-ac1d-cf0e925fef87 (1) (00:01:46)
diego-cell: diego-cell/003cb5af-debc-467c-b8f2-89ec1fb42ab1 (2) (00:01:39)
diego-cell: diego-cell/d18a2097-6263-491d-94ef-21e901de757c (1) (00:01:52)
diego-api: diego-api/c724a464-af90-4cd0-a30a-0f6c074580ba (0) (00:01:11)
diego-api: diego-api/80d61e49-9592-4c71-8269-0648b5e6600d (1) (00:01:01)
scheduler: scheduler/9effdf86-cf44-47dd-8c92-c1b4b02bd98e (0) (00:01:27)
scheduler: scheduler/4fb8b29a-ae03-45ef-9762-01133aa26f5a (1) (00:01:19)
log-api: log-api/062ed72d-61c5-4793-b364-53dd466ed8f8 (0) (00:01:09)
log-api: log-api/cb894008-88bd-45e9-b982-9f724341ff4a (1) (00:01:00)
cc-worker: cc-worker/35185635-c49f-4350-a45c-700469efc751 (1) (00:00:58)
```
