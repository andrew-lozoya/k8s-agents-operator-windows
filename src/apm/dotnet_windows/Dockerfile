# Use the Windows Nano Server for 2019 as the base image
FROM mcr.microsoft.com/windows/nanoserver:ltsc2019

# Argument to specify the version of the New Relic agent
ARG version

# Set the working directory
WORKDIR \instrumentation

# Download and extract the New Relic .NET agent
RUN curl.exe -L -o newrelic-agent.zip https://download.newrelic.com/dot_net_agent/latest_release/NewRelicDotNetAgent_%version%_x64.zip\
    && tar.exe -xzf newrelic-agent.zip\
    && del newrelic-agent.zip