#!/usr/bin/env python3

import argparse
import getpass
import os
import paramiko
import sys
import tarfile
import time
from io import BytesIO

def main():
    parser = argparse.ArgumentParser(description="Script de deploy (build remoto) no Docker Swarm/Portainer via SSH.")
    parser.add_argument("-H", "--host", help="Endereço IP/Host do servidor e porta (ex: 1.2.3.4:22)")
    parser.add_argument("-u", "--user", default="root", help="Usuário SSH (padrão: root)")
    parser.add_argument("-s", "--stack", help="Nome da stack no Docker Swarm (padrão: nome da pasta do projeto)")
    parser.add_argument("-p", "--project-dir", default=".", help="Diretório local do projeto a enviar (padrão: diretório atual)")
    parser.add_argument("--remote-dir", help="Diretório remoto onde o código será descompactado (padrão: /tmp/<nome_da_pasta_do_projeto>)")
    # Novo argumento para a tag personalizada
    parser.add_argument("--image-tag", default="latest", help="Tag da imagem Docker (padrão: latest)")
    args = parser.parse_args()

    # Determinar o diretório do projeto
    project_dir = args.project_dir
    abs_project_dir = os.path.abspath(project_dir)
    project_name = os.path.basename(abs_project_dir)

    # Definir stack e remote_dir dinamicamente, se não fornecidos
    stack_name = args.stack if args.stack else project_name
    remote_dir = args.remote_dir if args.remote_dir else f"/tmp/{project_name}"

    # Converter nomes para minúsculas
    stack_name = stack_name.lower()
    image_name = f"{stack_name}:{args.image_tag}".lower()

    # Se não informar host, pedir input
    host = args.host
    if not host:
        host = input("Informe o IP/Host do servidor + porta (ex: 1.2.3.4:22): ").strip()

    # Usuário SSH
    user = args.user
    if not user:
        user = input("Informe o usuário SSH (ex: root): ").strip()

    # Ler senha SSH
    password = getpass.getpass("Informe a senha SSH: ")

    ssh_client = None
    sftp_client = None
    try:
        # Conexão SSH
        ssh_client = connect_ssh(host, user, password)
        sftp_client = ssh_client.open_sftp()

        # 1) Gerar tar.gz do projeto local
        tar_data = create_project_tar(abs_project_dir)

        # 2) Enviar tar.gz para o servidor
        tar_filename = f"/tmp/{int(time.time())}_app.tar.gz"
        print(f"[INFO] Enviando projeto como {tar_filename} ...")
        upload_bytes_sftp(sftp_client, tar_data, tar_filename)
        print("[INFO] Arquivo .tar.gz enviado com sucesso.")

        # 3) Remover remote_dir anterior
        cmd_rm = f"rm -rf {remote_dir}"
        run_remote_command(ssh_client, cmd_rm)

        # 4) Criar pasta e descompactar
        cmd_extract = f"mkdir -p {remote_dir} && tar xzf {tar_filename} -C {remote_dir}"
        print("[INFO] Descompactando código no servidor...")
        run_remote_command(ssh_client, cmd_extract)

        # 5) Monta o nome da imagem com a tag
        # Já convertido para lowercase anteriormente
        # image_name = f"{stack_name}:{args.image_tag}".lower()

        # Faz build da imagem usando a nova tag
        # Observação: Se quiser DUAS tags (latest e a custom), use:
        #   docker build -t {stack_name}:latest -t {image_name} {remote_dir}
        cmd_build = f"docker build -t {image_name} {remote_dir}"
        print(f"[INFO] Executando build remoto: {cmd_build}")
        build_output, build_exit_status = run_remote_command(ssh_client, cmd_build)
        print("== Saída do build ==\n", build_output)

        # Verificar se o build foi bem-sucedido
        if build_exit_status != 0:
            print("[ERRO] O build falhou. Abortando deploy.")
            sys.exit(1)

        # 6) Deploy da stack
        compose_path = f"{remote_dir}/docker-compose.yml"
        cmd_deploy = f"docker stack deploy -c {compose_path} {stack_name}"
        print(f"[INFO] Realizando deploy: {cmd_deploy}")
        deploy_output, deploy_exit_status = run_remote_command(ssh_client, cmd_deploy)
        print("== Saída do deploy ==\n", deploy_output)

        # Opcional: Verificar status do deploy
        if deploy_exit_status != 0:
            print("[ERRO] O deploy falhou.")
            sys.exit(1)

        print("[SUCESSO] Deploy concluído com sucesso!")
    except Exception as e:
        print("[ERRO] Ocorreu um problema durante o deploy:", e)
        sys.exit(1)
    finally:
        if sftp_client:
            sftp_client.close()
        if ssh_client:
            ssh_client.close()

def connect_ssh(host, user, password):
    """
    Conecta no servidor SSH usando host:porta, usuário e senha.
    """
    ssh_client = paramiko.SSHClient()
    ssh_client.set_missing_host_key_policy(paramiko.AutoAddPolicy())

    if ":" in host:
        hostname, port_str = host.split(":")
        port = int(port_str)
    else:
        hostname = host
        port = 22

    print(f"[INFO] Conectando em {hostname}:{port} como {user}...")
    ssh_client.connect(hostname=hostname, port=port, username=user, password=password)
    return ssh_client

def create_project_tar(project_dir):
    """
    Cria um arquivo tar.gz em memória de todo o diretório local do projeto.
    Inclui Dockerfile, docker-compose.yml e outros arquivos.
    """
    buf = BytesIO()
    with tarfile.open(mode="w:gz", fileobj=buf) as tarf:
        for root, dirs, files in os.walk(project_dir):
            for f in files:
                fullpath = os.path.join(root, f)
                relpath = os.path.relpath(fullpath, project_dir)
                tarf.add(fullpath, arcname=relpath)
    buf.seek(0)
    return buf.read()

def upload_bytes_sftp(sftp_client, data_bytes, remote_path):
    """
    Envia bytes (data_bytes) para o caminho remoto (remote_path) via SFTP.
    """
    with sftp_client.open(remote_path, 'wb') as remote_file:
        remote_file.write(data_bytes)

def run_remote_command(ssh_client, command):
    """
    Executa um comando remoto e retorna stdout+stderr como string e o código de saída.
    """
    stdin, stdout, stderr = ssh_client.exec_command(command)
    out = stdout.read().decode("utf-8", errors="replace")
    err = stderr.read().decode("utf-8", errors="replace")
    exit_status = stdout.channel.recv_exit_status()
    return out + err, exit_status

# Necessário para que possamos rodar diretamente este script
if __name__ == "__main__":
    main()
