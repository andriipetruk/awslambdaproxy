import logging
from subprocess import Popen, PIPE

logger = logging.getLogger()
logger.setLevel(logging.INFO)

# Handler that will be called by Lambda
def handler(event, context):
    logger.info("Event: {}".format(event))
    logger.info("Context: {}".format(context))

    address = event['ConnectBackAddress']
    client_private_key = event['ClientPrivateKey']
    client_public_key = event['ClientPublicKey']
    server_public_key = event['ServerPublicKey']

    command = './awslambdaproxy-lambda -address="{}" -client-private-key="{}" -client-public-key="{}" ' \
              '-server-public-key="{}"'.format(address, client_private_key, client_public_key, server_public_key)
    logger.info("Running: {}".format(command))
    try:
        proc = Popen(command, shell=True, stdout=PIPE, stderr=PIPE)
        out, err = proc.communicate()
        print out, err, proc.returncode
    except Exception as e:
        logger.error("Error: {}".format(e))
