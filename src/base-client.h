#pragma once

#include "base.h"

class BaseClient : X32Base
{
    protected:
        ClientConfig* clientconfig;

    public:
        BaseClient(X32BaseParameter* basepar, ClientConfig cc) : X32Base(basepar)
        {
            clientconfig = cc;
        }
};