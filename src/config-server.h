#pragma once

#include "config.h"

namespace OMC
{
    class ConfigServer : Config
    {
        private:

           

        public:

            ConfigServer(String model, Helper* h);

            bool LoadConfig(uint scene);
            void Save(uint scene);
    };
}