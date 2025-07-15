# Wikipedia Todo Generator CronJob

This CronJob automatically generates a new todo item every hour with a random Wikipedia article to read.

## What it does

- Runs every hour (at minute 0 of each hour)
- Fetches a random Wikipedia article URL from https://en.wikipedia.org/wiki/Special:Random
- Creates a new todo with the text "Read <URL>" where <URL> is the random Wikipedia article
- Posts the todo to the todo-backend API

## Files

- `create-todo.sh` - Bash script that fetches random Wikipedia URL and creates the todo
- `Dockerfile` - Container image definition
- `../manifests/wikipedia-todo-cronjob.yaml` - Kubernetes CronJob manifest
- `build-and-deploy.sh` - Script to build and deploy the CronJob

## Usage

### 1. Build and Deploy

```bash
cd TheProject/wikipedia-cronjob
./build-and-deploy.sh
```

### 2. Check CronJob Status

```bash
kubectl get cronjobs -n project
kubectl get jobs -n project
```

### 3. Manually Trigger for Testing

```bash
kubectl create job --from=cronjob/wikipedia-todo-generator manual-test -n project
```

### 4. View Logs

```bash
# View logs from the most recent job
kubectl logs -l app=wikipedia-todo-generator -n project

# View logs from a specific job
kubectl logs job/manual-test -n project
```

### 5. View Generated Todos

Check your todo application - you should see new todos with "Read https://en.wikipedia.org/wiki/..." every hour.

## Schedule

The CronJob is configured to run every hour using the cron expression `0 * * * *`:
- `0` - At minute 0
- `*` - Every hour
- `*` - Every day of month
- `*` - Every month
- `*` - Every day of week

## Requirements

- The todo-backend service must be running in the `project` namespace
- The CronJob requires internet access to fetch random Wikipedia URLs
- The container needs `curl` and `bash` (provided by Alpine Linux base image) 