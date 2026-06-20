#pragma once

#include <algorithm>
#include <map>
#include <set>

#include "types.h"
#include "../lib/WString.h"

#include "helper.h"

#include "parameter.h"
#include "parameter.h"
#include "surfaceelement.h"
#include "x32faderbank.h"
#include "x32assignbank.h"

using enum MP_ID;
using namespace OMC;

class Config
{
    protected:

        Helper* helper;

        Parameter<MP_ID, MP_CAT>* mpm[(uint)MP_ID::__ELEMENT_COUNTER_DO_NOT_MOVE]; 
        
        map<MP_ID, set<uint>>* mp_changedlist = new map<MP_ID, set<uint>>();
        bool MixerParameterChangelistFreeze = false;
        map<MP_ID, set<uint>>* mp_changedlist_temp = new map<MP_ID, set<uint>>();
        
        void SetParameterChanged(MP_ID mp, uint index);

        // old
        OMC_MODEL _model;

    public:

        Config(String model, Helper* h);

        void DefineMixerparameters();
        Mixerparameter* DefParameter(MP_ID mp_type, MP_CAT category, String name, uint count = 1);
        Mixerparameter* GetParameter(MP_ID mp);

        Mixerparameter** GetParameterList();
        map<MP_ID, set<uint>>* GetChangedParameterList();

        MP_ID MpCalcId(MP_ID mp_id, int amount);
        
        vector<uint> GetChangedParameterIndexes(MP_CAT parameter_cat);
        vector<uint> GetChangedParameterIndexes(vector<MP_ID> filter_ids);
        bool HasParameterChanged(MP_ID parameter_id);
        bool HasParameterChanged(MP_ID parameter_id, uint index);
        bool HasParametersChanged(vector<MP_ID> parameter_id);
        bool HasParametersChanged(vector<MP_ID> parameter_id, uint index);
        bool HasParametersChanged(MP_CAT parameter_cat);
        bool HasParametersChanged(MP_CAT parameter_cat, uint index);
        bool HasAnyParameterChanged();
        void FreezeParameterList();
        void SetParameterUnchanged(MP_ID mp);
        void SaveResetAndUnfreezeChangedParameterList();
        
        float GetFloat(MP_ID mp, uint index = 0);
        int GetInt(MP_ID mp, uint index = 0);
        uint GetUint(MP_ID mp, uint index = 0);
        bool GetBool(MP_ID mp, uint index = 0);
        String GetString(MP_ID mp, uint index = 0);
        uint GetPercent(MP_ID mp, uint index = 0);
        bool GetBlink(MP_ID mp);
        void Set(MP_ID mp, float value, uint index = 0);    
        void Set(MP_ID mp, String value_string, uint index = 0
        );
        void Change(MP_ID mp, int amount, uint index = 0);
        void Toggle(MP_ID mp, uint index = 0);
        void Refresh(MP_ID mp, uint index = 0);
        void Reset(MP_ID mp, uint index = 0);

        MP_ID ParameterCalcId(SurfaceBindingParameter* binding_parameter);
        uint ParameterCalcIndex(SurfaceBindingParameter* binding_parameter);
        MP_ID ParameterDependsOn(SurfaceBindingParameter* binding_parameter);
        MP_ID ParameterDependsOn(MixerparameterAction mp_action);

        bool IsModelX32Full();
        bool IsModelX32FullOrM32();
        bool IsModelX32FullOrCompactOrM32();
        bool IsModelX32FullOrCompactOrProducerOrM32OrM32R();
        bool IsModelX32FullOrCompactOrProducerOrM32OrM32ROrRack();
        bool IsModelX32CompactOrProducerOrM32R();
        bool IsModelX32ProducerOrRackOrM32R();
        bool IsModelX32Core();
        bool IsModelX32Rack();
        bool IsModelX32Producer();
        bool IsModelX32Compact();
        bool IsModelX32CompactOrM32R();
        bool IsModelM32();
        bool IsModelM32R();
        bool IsModelM32C();
        bool IsModelAnyXM32();
        bool IsModelWingFull();
        bool IsModelWingCompact();
        bool IsModelWingRack();
        bool IsModelAnyWing();

        bool HasDisplay();
        bool HasBigDisplay();
        bool HasSmallDisplay();
        bool HasTouchDisplay();
};

class X32ConfigFileEntry
{
    public:
        MP_ID MixerparameterId;
        String MixerparameterName;
        vector<float> value;
        vector<String> string_value;
};

class X32ConfigFile
{
    public:
        vector<X32ConfigFileEntry> entries;
};

