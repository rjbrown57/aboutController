FROM ubuntu

RUN groupadd -r aboutController -g 99999 && \
    useradd -r -u 99999 -g aboutController -s /bin/bash aboutController

COPY aboutController /usr/local/bin/aboutController
RUN chown aboutController:aboutController /usr/local/bin/aboutController && \
    chmod 755 /usr/local/bin/aboutController

USER aboutController

ENTRYPOINT ["/usr/local/bin/aboutController"]
