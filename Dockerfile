FROM ubuntu:21.04

ARG USER="user"
ARG PASS="test"

# TODO: --password is insecure
RUN useradd --no-log-init -ms /bin/bash ${USER} \
						--password ${PASS}

#RUN echo $USER'${USER}:${PASS}' | chpasswd

USER ${USER}

WORKDIR /home/${USER}

CMD ["/bin/bash"]
