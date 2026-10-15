#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

export SAM_CLI_TELEMETRY=0
export AWS_PAGER=""

for cmd in aws sam docker; do
	if ! command -v "$cmd" >/dev/null 2>&1; then
		echo "Falta el comando '$cmd'." >&2
		exit 1
	fi
done

if [[ -f .env ]]; then
	set -a
	# shellcheck disable=SC1091
	source .env
	set +a
fi

STACK_NAME="${STACK_NAME:-spin-transactions}"
REGION="${AWS_REGION:-${AWS_DEFAULT_REGION:-}}"
if [[ -z "$REGION" ]]; then
	REGION="$(aws configure get region || true)"
fi
if [[ -z "$REGION" ]]; then
	echo "Define AWS_REGION o una región en el perfil de AWS." >&2
	exit 1
fi

PROVIDER_BASE_URL="${PROVIDER_BASE_URL:-https://em4d4.wiremockapi.cloud}"
PROVIDER_API_KEY="${PROVIDER_API_KEY:-demo-api-key}"
IMAGE_TAG="${IMAGE_TAG:-$(date -u +%Y%m%d%H%M%S)}"

aws sts get-caller-identity >/dev/null

echo "Stack $STACK_NAME en $REGION. Crea RDS, un ALB y dos tasks Fargate: tiene costo y el primer deploy tarda varios minutos."

sam_deploy() {
	local deploy_services="$1"
	local image_tag="$2"
	sam deploy \
		--stack-name "$STACK_NAME" \
		--region "$REGION" \
		--resolve-s3 \
		--capabilities CAPABILITY_IAM \
		--no-confirm-changeset \
		--no-fail-on-empty-changeset \
		--parameter-overrides \
		"ParameterKey=DeployServices,ParameterValue=${deploy_services}" \
		"ParameterKey=ImageTag,ParameterValue=${image_tag}" \
		"ParameterKey=ProviderBaseUrl,ParameterValue=${PROVIDER_BASE_URL}" \
		"ParameterKey=ProviderApiKey,ParameterValue=${PROVIDER_API_KEY}"
}

stack_status=""
if stack_status="$(aws cloudformation describe-stacks \
	--stack-name "$STACK_NAME" \
	--region "$REGION" \
	--query "Stacks[0].StackStatus" \
	--output text 2>/dev/null)"; then
	case "$stack_status" in
	CREATE_COMPLETE | UPDATE_COMPLETE | UPDATE_ROLLBACK_COMPLETE) ;;
	ROLLBACK_COMPLETE)
		echo "El stack quedó en ROLLBACK_COMPLETE. Bórralo y vuelve a correr make deploy." >&2
		exit 1
		;;
	*)
		echo "El stack está en $stack_status. Espera a que termine antes de volver a desplegar." >&2
		exit 1
		;;
	esac
else
	echo "Creando red, RDS, ECR y el balanceador."
	sam_deploy false latest
fi

image_uri="$(aws cloudformation describe-stacks \
	--stack-name "$STACK_NAME" \
	--region "$REGION" \
	--query "Stacks[0].Outputs[?OutputKey=='EcrRepositoryUri'].OutputValue" \
	--output text)"
if [[ -z "$image_uri" || "$image_uri" == "None" ]]; then
	echo "El stack no publicó EcrRepositoryUri." >&2
	exit 1
fi

registry="${image_uri%%/*}"
aws ecr get-login-password --region "$REGION" | docker login --username AWS --password-stdin "$registry"

echo "Publicando $image_uri:$IMAGE_TAG"
docker build --platform linux/arm64 --provenance=false -t "$image_uri:$IMAGE_TAG" .
docker push "$image_uri:$IMAGE_TAG"

echo "Desplegando los servicios write y read."
sam_deploy true "$IMAGE_TAG"

alb="$(aws cloudformation describe-stacks \
	--stack-name "$STACK_NAME" \
	--region "$REGION" \
	--query "Stacks[0].Outputs[?OutputKey=='LoadBalancerDNS'].OutputValue" \
	--output text)"
echo "POST http://${alb}/transactions"
echo "GET  http://${alb}/transactions"
