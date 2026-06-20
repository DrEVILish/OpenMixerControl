#pragma once

#include "x32config.h"
#include "state.h"
#include "helper.h"
#include "base-parameter.h"

class X32Base
{
    protected:
        CLI::App* app;
        ServerConfig* config;
        State* state;
        Helper* helper;

    public:
        X32Base(X32BaseParameter* basepar);
};