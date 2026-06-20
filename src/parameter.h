#pragma once

#include <stdint.h>
#include <vector>
#include "../lib/WString.h"
#include "glaze/glaze.hpp"
#include "math.h"

#include "defines.h"
#include "enum.h"

using namespace std;
using namespace WString;

namespace OMC
{
    template <typename T_ID, typename T_CAT> class Parameter
    {
        friend struct glz::meta<Parameter<T_ID, T_CAT>>;

        private:

            T_ID parameter_id;
            T_CAT category;        // category of parameter, e.g. setting, channel, fx
            MP_VALUE_TYPE value_type;

            String _name;           // complete long name
            String _name_short;     // short name for thight spaces

            bool _no_config = false;  // include Mixerparameter NOT in saved config

            MP_UOM unitOfMeasurement = MP_UOM::NONE;

            float value_min;
            float value_max;
            float value_standard;

            String value_string_standard;

            /// @brief How many instances this Mixerparameter has.
            uint instances = 0;
            
            vector<float> value;
            vector<String> value_string;

            uint decimal_places = 0;
            uint stepmode = 0; // 0: linear, 1: frequency, 2: boolean toggle between 0 and 1
            
            // When the resulting value is too high:
            // 0: limit to max value
            // 1: set to min value
            uint cyclemode_high = 0; 
            
            // When the resulting value is too low:
            // 0: limit to min value
            // 1: set to max value
            uint cyclemode_low = 0;

            float stepsize = 1;

            bool value_string_autoincrement_zerobased = false;

            bool button_blink = false;

            bool readonly = false;

            MixerparameterAction defaultAction = MixerparameterAction::NONE;

            /// @brief assign members via select if this Mixerparameter is true
            T_ID assign_members_if = T_ID::NONE;
            /// @brief assign members via select to this Mixerparameter
            T_ID assign_members_to = T_ID::NONE;

            /// @brief Hide slider on display encoder.
            bool hide_encoder_slider = false;        

            /// @brief Hide reset label and function on display encoder.
            bool hide_encoder_reset = false;  

            String osc_path = "";

            /// @brief Checks if the index is within the specified size of the Mixerparameter.
            /// @param index The index to check. 
            /// @throws std::out_of_range
            void CheckIndex(uint index)
            {
                if (index >= instances)
                {
                    __throw_out_of_range((String("The index ") + String(index) + String(" is bigger than the specified instances of ") + String(instances) + String(" (zero based!) of the Mixerparameter ") + GetName() + String(".")).c_str());
                }
            }

            /// @brief Checks if the datatype is the datatype of the Mixerparameter.
            /// @param mp_value_type The datatype to check. 
            /// @throws std::bad_typeid
            void CheckDatatype(MP_VALUE_TYPE mp_value_type)
            {
                if (mp_value_type != value_type)
                {
                    __throw_bad_typeid();
                }
            }

            /// @brief Checks if the Mixerparameter is not readonly and his data can be changed.
            /// @throws std::logic_error
            void CheckNotReadonly()
            {
                if (readonly)
                {
                    __throw_logic_error((String("The Mixerparameter ") + GetName() + String(" can not be changed, it is readonly.")).c_str());
                }
            }

            String GetUnitOfMesaurement(bool spaceInFront, uint index, bool isResetLabel);
            String FormatValue_Intern(float value_float, uint index, bool isResetLabel);

        public:
    
            Parameter(T_ID mp, T_CAT cat, String name, uint count)
            {
                parameter_id = mp;
                category = cat;

                _name = name;

                DefNameShort(name);

                instances = count;
            }

            // Short name of Mixerparameter, max. 5 Characters!
            Parameter* DefNameShort(String name_short) 
            {
                _name_short = name_short;
                if (_name_short.length() > 5) {
                    _name_short = _name_short.substring(0, 5);
                }

                return this;
            }

            Parameter* DefUOM(MP_UOM uom) 
            {
                unitOfMeasurement = uom;

                switch (uom) {
                    using enum MP_UOM;

                    case PERCENT:
                        stepsize = 0.01f;  // percent in 1% steps
                        decimal_places = 1;
                        break;
                    case SECONDS:
                        stepsize = 0.2f;  // seconds in 200ms steps
                        break;
                    default:
                        stepsize = 1.0f;
                        break;
                }

                return this;
            }

            Parameter* DefMinMaxStandard_Float(float min, float max, float standard, uint decimals = 0) {
                value_type = MP_VALUE_TYPE::FLOAT;
                value_min = min;
                value_max = max;
                value_standard = standard;
                decimal_places = decimals;

                // fill with default values
                for (uint i = 0; i < instances; i++)
                {
                    value.push_back(value_standard);
                }

                return this;
            }

            Parameter* DefMinMaxStandard_Uint(uint min, uint max, uint standard) {
                value_type = MP_VALUE_TYPE::UINT;       
                value_min = min;
                value_max = max;
                value_standard = standard;
                decimal_places = 0;

                // fill with default values
                for (uint i = 0; i < instances; i++)
                {
                    value.push_back(value_standard);
                }

                return this;
            }

            Parameter* DefMinMaxStandard_Int(int min, int max, int standard) {
                value_type = MP_VALUE_TYPE::INT;
                value_min = min;
                value_max = max;
                value_standard = standard;
                decimal_places = 0;

                // fill with default values
                for (uint i = 0; i < instances; i++)
                {
                    value.push_back(value_standard);
                }

                return this;
            }

            Parameter* DefStandard_Bool(bool standard) {
                value_type = MP_VALUE_TYPE::BOOL;
                value_min = 0;
                value_max = 1;
                value_standard = standard;
                decimal_places = 0;
                hide_encoder_slider = true;

                // fill with default values
                for (uint i = 0; i < instances; i++)
                {
                    value.push_back(value_standard);
                }

                return this;
            }

            Parameter* DefStandard_String(String standard, bool autoincrement_zerobased = false) {
                value_type = MP_VALUE_TYPE::STRING;
                
                value_string_standard = standard;
                value_string_autoincrement_zerobased = autoincrement_zerobased;
                hide_encoder_slider = true;
                hide_encoder_reset = true;

                // fill with default values
                for (uint i = 0; i < instances; i++)
                {
                    value_string.push_back(value_string_standard + String(autoincrement_zerobased ? i : i+1));
                }

                return this;
            }

            Parameter* DefStepsize(float steps) {
                stepsize = steps;

                return this;
            }

            Parameter* DefStepmode(uint mode) {
                stepmode = mode;

                return this;
            }

            // When the resulting value is too low:
            // 0: set to min value (limit)
            // 1: set to max value (cycle around)
            // When the resulting value is too high:
            // 0: set to max value (limit)
            // 1: set to min value (cycle around)
            Parameter* DefCycleMode(uint mode_low, uint mode_high)
            {
                cyclemode_low = mode_low;
                cyclemode_high = mode_high;

                return this;
            }

            // If the Mixerparameter is bound to a button and its value is true, the button blinks.
            Parameter* DefButtonBlink()
            {   
                button_blink = true;

                return this;
            }

            /// @brief The data of this Mixerparameter can not be changed, it is readonly.
            Parameter* DefReadonly() {
                readonly = true;

                return this;
            }

            /// @brief assign members via select if this Mixerparameter is true
            Parameter* DefAssignMembersIfTo(T_ID MixerparameterToCheck, T_ID MixerparameterDestination)
            {
                assign_members_if = MixerparameterToCheck;
                assign_members_to = MixerparameterDestination;

                return this;
            }

            /// @brief Hide the slider on display encoders if this Mixerparameter is bound.
            Parameter* DefHideEncoderSlider() {
                hide_encoder_slider = true;

                return this;
            }

            /// @brief Hide the reset on display encoders if this Mixerparameter is bound.
            Parameter* DefHideEncoderReset() {
                hide_encoder_reset = true;

                return this;
            }

            /// @brief This Mixerparameter in NOT included in the configfile.
            Parameter* DefNoConfig()
            {
                _no_config = true;

                return this;
            }

            Parameter* DefOSC(String path)
            {
                osc_path = path;

                return this;
            }

            //############################################################################################

            /// @brief Get the ID of the Mixerparameter.
            /// @return Mixerparameter ID. 
            T_ID GetId()
            {
                return parameter_id;
            }

            /// @brief Get the data type of the stored value.
            /// @return The data type.
            MP_VALUE_TYPE GetType()
            {
                return value_type;
            }

            bool GetHideEncoderSlider()
            {
                return hide_encoder_slider;
            }


            bool GetHideEncoderReset()
            {
                return hide_encoder_reset;
            }


            /// @brief Get the category of the Mixerparameter.
            /// @return The category.
            T_CAT GetCategory()
            {
                return category;
            }

            T_ID GetAssignMembersIf()
            {
                return assign_members_if;
            }

            T_ID GetAssignMembersTo()
            {
                return assign_members_to;
            }

            uint GetDecimaPlaces()
            {
                CheckDatatype(MP_VALUE_TYPE::FLOAT);

                return decimal_places;
            }

            MP_UOM GetUOM()
            {
                return unitOfMeasurement;
            }

            /// @brief Resets the stored data.
            /// @param index The index of the parameter (usual the vchannel index or FX slot index).
            void Reset(uint index = 0)
            {
                CheckIndex(index);
                CheckNotReadonly();
                
                if (value_type == MP_VALUE_TYPE::STRING)
                {
                    value_string[index] = value_string_standard + String(index);
                }
                else
                {
                    value[index] = value_standard;
                }
            }

            /// @brief Sets a new value to the Mixerparameter.
            /// @param newValue The new value.
            /// @param index The index of the parameter (usual the vchannel index or FX slot index).
            void Set(float newValue, uint index = 0)
            {
                CheckIndex(index);
                CheckNotReadonly();

                if (newValue > value_max) {
                    newValue = value_max;
                } else if (newValue < value_min) {
                    newValue = value_min;
                }
                value[index] = newValue;
            }

            /// @brief Sets a new value to the Mixerparameter.
            /// @param newValue The new value.
            /// @param index The index of the parameter (usual the vchannel index or FX slot index).
            void Set(uint newValue, uint index = 0)
            {
                CheckIndex(index);
                CheckNotReadonly();

                if (newValue > value_max) {
                    newValue = value_max;
                } else if (newValue < value_min) {
                    newValue = value_min;
                }
                value[index] = newValue;
            }

            /// @brief Sets a new value to the Mixerparameter.
            /// @param newValue The new value.
            /// @param index The index of the parameter (usual the vchannel index or FX slot index).
            void Set(int newValue, uint index = 0)
            {
                CheckIndex(index);
                CheckNotReadonly();

                if (newValue > value_max) {
                    newValue = value_max;
                } else if (newValue < value_min) {
                    newValue = value_min;
                }
                value[index] = newValue;
            }

            /// @brief Sets a new value to the Mixerparameter.
            /// @param newValue The new value.
            /// @param index The index of the parameter (usual the vchannel index or FX slot index).
            void Set(bool newValue, uint index = 0)
            {
                CheckIndex(index);
                CheckNotReadonly();

                if (newValue > value_max) {
                    newValue = value_max;
                } else if (newValue < value_min) {
                    newValue = value_min;
                }
                value[index] = newValue;
            }

            /// @brief Sets a new value to the Mixerparameter.
            /// @param newValue The new value.
            /// @param index The index of the parameter (usual the vchannel index or FX slot index).
            void Set(String newValue, uint index = 0)
            {
                CheckIndex(index);
                CheckNotReadonly();

                value_string[index] = newValue;
            }

            float Get(uint index = 0)
            {
                return GetFloat(index);
            }

            float GetFloat(uint index = 0)
            {
                if (index < instances)
                {
                    return value[index];
                }

                return 0.0f;
            }

            int GetInt(uint index = 0)
            {
                if (index < instances)
                {            
                    return (int)value[index];
                }

                return 0;
            }

            uint GetUint(uint index = 0)
            {
                if (index < instances)
                {
                    return (uint)value[index];
                }

                return 0;
            }

            String GetString(uint index = 0)
            {
                if (index < instances && value_type == MP_VALUE_TYPE::STRING)
                {
                    return value_string[index];
                }

                return "";
            }

            uint GetPercent(uint index = 0)
            {
                if (index < instances)
                {
                    float onehunderedpercent = value_max - value_min;
                    float value_normiert = value[index] - value_min;
                    float onepercent = onehunderedpercent / 100.0f;
                    return (uint)(value_normiert / onepercent);
                }

                return 0;
            }

            bool GetBool(uint index = 0)
            {
                if (index < instances && value_type == MP_VALUE_TYPE::BOOL)
                {
                    return (bool)value[index];
                }

                return false;
            }

            float GetMin()
            {
                return value_min;
            }

            float GetMax()
            {
                return value_max;
            }

            bool GetBlink()
            {
                return button_blink;
            }

            void Change(int amount, uint index = 0)
            {
                CheckIndex(index);
                CheckNotReadonly();

                // Strings can not be 'changed'
                if (value_type == MP_VALUE_TYPE::STRING)
                {
                    return;
                }

                if (stepsize == 0)
                {
                    __throw_logic_error((String("Stepsize of Mixerparameter ") + GetName() + String(" is 0, so no change can happen!")).c_str());
                }

                float newValue;
                switch (stepmode)
                {
                    case 0: // linear
                        newValue = value[index] + (amount * stepsize);
                        break;
                    case 1: // frequency
                        newValue = value[index] * powf(2.0f, (amount * stepsize) / 12.0f); // semitone steps
                        break;
                    case 2: // boolean toggle
                        newValue = (amount > 0) ? 1.0f : 0.0f;
                        break;
                    default:
                        newValue = value[index] + (amount * stepsize);
                        break;
                }

                // resulting value is too low
                if (newValue < value_min)
                {
                    switch (cyclemode_low)
                    {
                        case 0: // limit to min value
                            newValue = value_min;
                            break;
                        case 1:
                            newValue = value_max;
                            break;
                    }
                }

                // resulting value is too high
                if (newValue > value_max)
                {
                    switch (cyclemode_high)
                    {
                        case 0:
                            newValue = value_max;
                            break;
                        case 1:
                            newValue = value_min;
                            break;
                    }
                }
                
                Set(newValue, index);            
            }

            bool Toggle(uint index = 0)
            {
                CheckIndex(index);
                CheckNotReadonly();
                
                if (value_type == MP_VALUE_TYPE::BOOL)
                {
                    // toggle between true and false
                    bool bool_value = value[index] != 0.0f;
                    value[index] = !bool_value;
                }else if (value_type == MP_VALUE_TYPE::FLOAT || value_type == MP_VALUE_TYPE::UINT || value_type == MP_VALUE_TYPE::INT)
                {
                    // increase value until max is reached, then set to min
                    if (value[index] < value_max)
                    {
                        if (stepsize > 0) {
                            // increase by set stepsize
                            Set(value[index] + stepsize, index);
                        }else{
                            // increase by 5%
                            Set(value[index] + ((value_max - value_min) * 0.05f), index);
                        }
                    }else{
                        // wrap around to min
                        Set(value_min, index);
                    }
                }else{
                    // Strings or other types are not supported here
                }
                return false;
            }

            /// @brief How many instances (Channels or FX-Slots) this parameter has.
            /// @return The number of instances.
            uint GetInstances()
            {        
                return instances;
            }


            void SetName(String name)
            {
                _name = name;
            }

            String GetName()
            {
                return _name;
            }

            String GetNameShort()
            {
                return _name_short;
            }

            String GetLabelAndValue(uint index = 0)
            {
                if (index < instances)
                {
                    return GetName() + String(": ") + GetFormatedValue(index);
                }

                return "";
            }

            String GetShortLabelAndValue(uint index = 0)
            {
                if (index < instances)
                {
                return GetNameShort() + String(": ") + GetFormatedValue(index);
                }

                return "";
            }

            String GetResetLabel(uint index = 0)
            {
                if (index < instances)
                {
                    return String("Reset: ") + FormatValue_Intern(value_standard, index, true);
                }

                return "";
            }

            String GetFormatedValue(uint index = 0)
            {
                if (index < instances)
                {
                    if (value_type == MP_VALUE_TYPE::STRING)
                    {
                        return value_string[index];
                    }
                    else
                    {
                        return FormatValue_Intern(value[index], index, false);
                    }
                }
                
                return "";
            }

            bool BelongsToChannel()
            {
                return
                    category == T_CAT::CHANNEL || 
                    category == T_CAT::CHANNEL_DYNAMICS ||
                    category == T_CAT::CHANNEL_EQ ||
                    category == T_CAT::CHANNEL_GATE ||
                    category == T_CAT::CHANNEL_SENDS;
            }

            bool BelongsToFX()
            {
                return category == T_CAT::FX;
            }

            MixerparameterAction GetPreferredAction()
            {
                MixerparameterAction action;

                switch (value_type)
                {
                    case MP_VALUE_TYPE::BOOL:
                        action = MixerparameterAction::TOGGLE;
                        break;
                    case MP_VALUE_TYPE::FLOAT:
                    case MP_VALUE_TYPE::UINT:
                    case MP_VALUE_TYPE::INT:
                        action = MixerparameterAction::CHANGE;
                        break;
                    case MP_VALUE_TYPE::STRING:
                        action = MixerparameterAction::NONE;
                        break;
                }

                return action;
            }

            /// @brief Get wether this Mixerparameter must not be in configfile.
            /// @return True if this Mixerparameter must not be in configfile. 
            bool IsNoConfig()
            {            
                return _no_config;
            }

            bool IsReadonly()
            {
                return readonly;
            }

            vector<float> Config_GetValue()
            {
                return value;
            }

            vector<String> Config_GetValueString()
            {
                return value_string;
            }

            void Config_SetValue(vector<float> newvalues)
            {
                value.clear();
                copy(newvalues.begin(), newvalues.end(), back_inserter(value));
            }

            void Config_SetValueString(vector<String> newvalues_string)
            {
                value_string.clear();
                copy(newvalues_string.begin(), newvalues_string.end(), back_inserter(value_string));
            }

            bool IsOSC()
            {
                return osc_path.length() > 0;
            }

            String GetOSC()
            {
                return osc_path;
            }
    };
}

template <typename T_ID, typename T_CAT> struct glz::meta<OMC::Parameter<T_ID, T_CAT>>
{
    using T = OMC::Parameter<T_ID, T_CAT>;
    static constexpr auto value = glz::object(
            &T::parameter_id,
            &T::value //,            &T::value_string
        );
};

namespace glz
{
    template <>
    struct from<JSON, String>
    {
        template <auto Opts>
        static void op(String& value, is_context auto&& ctx, auto&& it, auto&& end)
        {
            std::string str;
            parse<JSON>::op<Opts>(str, ctx, it, end);

            value = str.c_str();
        }
    };

    template <>
    struct to<JSON, String>
    {
        template <auto Opts>
        static void op(String& value, is_context auto&& ctx, auto&& b, auto&& ix) noexcept
        {
            serialize<JSON>::op<Opts>(value.c_str(), ctx, b, ix);
        }
    };
}