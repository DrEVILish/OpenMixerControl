#pragma once

#include "../lib/CLI11.hpp"
#include "x32config.h"
#include "state.h"
#include "helper.h"

class X32BaseParameter{
    public:
        CLI::App* app;
        ServerConfig* config;    
        State* state;
        Helper* helper;

        X32BaseParameter(CLI::App* a, ServerConfig* c, State* s, Helper* h){
            this->app = a;
            this->config = c;
            this->state = s;
            this->helper = h;
        }
};