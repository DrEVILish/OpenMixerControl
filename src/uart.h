#pragma once

#include <cstdint>
#include <stdio.h>
#include <string.h>
#include <linux/input.h>
#include <fcntl.h>
#include <termios.h>
#include <sys/ioctl.h> // for FIONREAD
#include <unistd.h>

#include "base.h"
#include "message-base.h"
#include "types.h"

class Uart : public X32Base
 {
    
    private:
        int fd = -1; // default: not connected
        bool force = false; // talk to the port even in bodyless mode

    public:
        Uart(X32BaseParameter* basepar);
        int Open(const char* ttydev, uint32_t baudrate, bool raw, bool force = false);
        int Tx(MessageBase* message);
        int Rx(char* buf, uint16_t bufLen);
        //void MirrorBack();
        void FlushRxBuffer();
};