# TikLab Sandbox - Base image
# Multi-stage: runtime based on evilfreelancer/docker-routeros:7
# Engine build stage added in Phase 5

FROM evilfreelancer/docker-routeros:7

# Copy entrypoint script
COPY scripts/entrypoint.sh /usr/local/bin/tiklab-entrypoint.sh
RUN chmod +x /usr/local/bin/tiklab-entrypoint.sh

# Expose ports: SSH, API, Winbox, Control
EXPOSE 22 8728 8291 9090

# Use custom entrypoint (Phase 5 will add engine binary)
ENTRYPOINT ["/usr/local/bin/tiklab-entrypoint.sh"]
