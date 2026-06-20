#pragma once

#include "config.h"

namespace OMC
{
    class ConfigClient : Config
    {
        private:

            // Surfaceelements
            SurfaceElement* sem[(uint)SurfaceElementId::__ELEMENT_COUNTER_DO_NOT_MOVE];

            // surface binding
            map<SurfaceElementId, SurfaceBindingParameter*>* surface_binding;
            set<SurfaceElementId> surface_binding_changed;

            OMCAssignBank* assingBanks[(uint)X32AssignBankId::__ELEMENT_COUNTER_DO_NOT_MOVE];

        public:

            ConfigClient(String model, Helper* h);

            void DefineSurfaceElements();
            SurfaceElement* DefSurfaceElements(SurfaceElementId element_id, String name);

            SurfaceElementId CalcSurfaceElementId(SurfaceElementId id, int amount);  

            bool HasSurfaceElement(SurfaceElementId id);
            SurfaceElement* GetSurfaceElement(SurfaceElementId);
            SurfaceElement* GetSurfaceElementButton_XM32(OMC_BOARD board, uint16_t value);
            SurfaceElement* GetSurfaceElementButton_Wing(OMC_BOARD board, uint index);
            SurfaceElement* GetSurfaceElementEncoder(OMC_BOARD board, uint8_t index);
            SurfaceElement* GetSurfaceElementFader(OMC_BOARD board, uint8_t index);
            
            map<SurfaceElementId, SurfaceBindingParameter*>* GetSurfaceBinding();
            SurfaceBindingParameter* GetSurfaceBinding(SurfaceElementId elementId);
            void SurfaceBindParameter(SurfaceElementId surfaceelement_id, SurfaceBindingParameter* binding_parameter);
            void SurfaceBind(SurfaceElementId surfaceelement_id, MixerparameterAction action, MP_ID mixerparaemter_id, uint mixerparameter_index = 0, uint extra_value = 0);
            // void SurfaceBind(SurfaceElementId surfaceelement_id, X32Action action);
            void SurfaceUnbind(SurfaceElementId surfaceelement_id);
            void SurfaceBindCustom(SurfaceElementId surfaceelement_id, String labeltext = "");

            bool HasAnySurfaceBindingChanged();
            bool HasSurfaceBindingChanged(SurfaceElementId elementId);
            void RemoveSurfaceBindingChanged(SurfaceElementId elementId);
            bool HasBoundParameterChanged(SurfaceElementId id);

            void InitAssignBanks();
            OMCAssignBank* GetAssignBank(X32AssignBankId id);
    };
}