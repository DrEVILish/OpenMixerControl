#pragma once

#include "config-server.h"
#include "config-client.h"
#include "ctrl-server.h"
#include "ctrl-client.h"

class OpenMixerControl
{
    private:

        bool runAsClient = false;

        CtrlServer* server = 0;
        CtrlClient* client = 0;

    public:

        OpenMixerControl(X32BaseParameter* basepar, ConfigServer cs, ConfigClient cc);
        
        void Init();
        
        void Tick10ms(void);
        void Tick50ms(void);
        void Tick100ms(void);

        void SimulatorButton();
};