FROM node:22-bookworm

RUN apt-get update \
    && apt-get install -y --no-install-recommends openjdk-17-jdk-headless unzip \
    && rm -rf /var/lib/apt/lists/*

ENV JAVA_HOME=/usr/lib/jvm/java-17-openjdk-amd64 \
    ANDROID_HOME=/opt/android-sdk
ENV PATH=$PATH:$ANDROID_HOME/cmdline-tools/latest/bin:$ANDROID_HOME/platform-tools

# Gradle downloads the SDK packages it needs once the licences are accepted
ARG ANDROID_CMDLINE_TOOLS=13114758
RUN mkdir -p $ANDROID_HOME/cmdline-tools \
    && curl -fsSL -o /tmp/tools.zip "https://dl.google.com/android/repository/commandlinetools-linux-${ANDROID_CMDLINE_TOOLS}_latest.zip" \
    && unzip -q /tmp/tools.zip -d $ANDROID_HOME/cmdline-tools \
    && mv $ANDROID_HOME/cmdline-tools/cmdline-tools $ANDROID_HOME/cmdline-tools/latest \
    && rm /tmp/tools.zip \
    && yes | sdkmanager --licenses > /dev/null

WORKDIR /app

# Dependencies
COPY package.json package-lock.json* ./
RUN npm install --no-audit --no-fund
COPY android ./android
RUN cd android && ./gradlew help --no-daemon -q

# Build (arm64 for phones, x86_64 for emulators)
COPY . .
RUN cd android \
    && ./gradlew assembleRelease --no-daemon -PreactNativeArchitectures=arm64-v8a,x86_64 \
    && cp app/build/outputs/apk/release/app-release.apk /opt/client.apk

CMD ["sh", "-c", "mkdir -p /out && cp /opt/client.apk /out/client.apk && echo 'APK written to ./dist/client.apk'"]
