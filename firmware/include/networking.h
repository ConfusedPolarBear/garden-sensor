#pragma once

void connectToWifi();
void startAccessPoint(int channel);
void stopAccessPoint();
void startNetworkScan();
void processNetworkScan();
void sendDiscoveryMessage(bool useMqtt);
