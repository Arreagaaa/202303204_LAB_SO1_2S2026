FROM alpine:3.20
RUN printf '#!/bin/sh\necho "lectura de prueba: /etc/hostname"\ncat /etc/hostname\nsleep 240\n' > /usr/local/bin/intruder.sh && chmod +x /usr/local/bin/intruder.sh
CMD ["/usr/local/bin/intruder.sh"]
