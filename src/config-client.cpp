#include "config-client.h"


ConfigClient::ConfigClient(String model, Helper* h) : Config(model, h)
{
    surface_binding = new map<SurfaceElementId, SurfaceBindingParameter*>();

    DefineSurfaceElements();
    InitAssignBanks();
}
    

void ConfigClient::InitAssignBanks()
{
    assingBanks[(uint)X32AssignBankId::Bank_A] = new OMCAssignBank(X32AssignBankId::Bank_A, String("Assign A"));
    assingBanks[(uint)X32AssignBankId::Bank_B] = new OMCAssignBank(X32AssignBankId::Bank_B, String("Assign B"));
    assingBanks[(uint)X32AssignBankId::Bank_C] = new OMCAssignBank(X32AssignBankId::Bank_C, String("Assign C"));

    if(IsModelX32Full())
	{
		OMCAssignBank* bank = assingBanks[(uint)X32AssignBankId::Bank_A];

		bank->bindingMap->at(SurfaceElementId::ASSIGN_ENCODER_1)->FillBindingParameter(MixerparameterAction::CHANGE, CHANNEL_VOLUME, 0);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_ENCODER_2)->FillBindingParameter(MixerparameterAction::CHANGE, CHANNEL_VOLUME, 1);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_ENCODER_3)->FillBindingParameter(MixerparameterAction::CHANGE, CHANNEL_VOLUME, 2);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_ENCODER_4)->FillBindingParameter(MixerparameterAction::CHANGE, CHANNEL_VOLUME, 3);

		bank->bindingMap->at(SurfaceElementId::ASSIGN_LCD_1)->FillBindingParameter(MixerparameterAction::LCD_Channel, NONE, 0);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_LCD_2)->FillBindingParameter(MixerparameterAction::LCD_Channel, NONE, 1);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_LCD_3)->FillBindingParameter(MixerparameterAction::LCD_Channel, NONE, 2);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_LCD_4)->FillBindingParameter(MixerparameterAction::LCD_Channel, NONE, 3);

		bank->bindingMap->at(SurfaceElementId::ASSIGN_5)->FillBindingParameter(MixerparameterAction::TOGGLE, CHANNEL_SOLO, 0);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_6)->FillBindingParameter(MixerparameterAction::TOGGLE, CHANNEL_SOLO, 1);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_7)->FillBindingParameter(MixerparameterAction::TOGGLE, CHANNEL_SOLO, 2);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_8)->FillBindingParameter(MixerparameterAction::TOGGLE, CHANNEL_SOLO, 3);

        bank->bindingMap->at(SurfaceElementId::ASSIGN_9)->FillBindingParameter(MixerparameterAction::TOGGLE, CHANNEL_MUTE, 0);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_10)->FillBindingParameter(MixerparameterAction::TOGGLE, CHANNEL_MUTE, 1);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_11)->FillBindingParameter(MixerparameterAction::TOGGLE, CHANNEL_MUTE, 2);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_12)->FillBindingParameter(MixerparameterAction::TOGGLE, CHANNEL_MUTE, 3);

        bank = assingBanks[(uint)X32AssignBankId::Bank_B];

        bank->bindingMap->at(SurfaceElementId::ASSIGN_ENCODER_1)->FillBindingParameter(MixerparameterAction::CHANGE, CHANNEL_VOLUME, 0);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_ENCODER_2)->FillBindingParameter(MixerparameterAction::CHANGE, CHANNEL_VOLUME, 1);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_ENCODER_3)->FillBindingParameter(MixerparameterAction::CHANGE, CHANNEL_VOLUME, 2);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_ENCODER_4)->FillBindingParameter(MixerparameterAction::CHANGE, CHANNEL_VOLUME, 3);

		bank->bindingMap->at(SurfaceElementId::ASSIGN_LCD_1)->FillBindingParameter(MixerparameterAction::LCD_Assign, NONE, 0);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_LCD_2)->FillBindingParameter(MixerparameterAction::LCD_Assign, NONE, 1);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_LCD_3)->FillBindingParameter(MixerparameterAction::LCD_Assign, NONE, 2);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_LCD_4)->FillBindingParameter(MixerparameterAction::LCD_Assign, NONE, 3);

		bank->bindingMap->at(SurfaceElementId::ASSIGN_5)->FillBindingParameter(MixerparameterAction::TOGGLE, CHANNEL_SOLO, 0);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_6)->FillBindingParameter(MixerparameterAction::TOGGLE, CHANNEL_SOLO, 1);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_7)->FillBindingParameter(MixerparameterAction::TOGGLE, CHANNEL_SOLO, 2);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_8)->FillBindingParameter(MixerparameterAction::TOGGLE, CHANNEL_SOLO, 3);

        bank->bindingMap->at(SurfaceElementId::ASSIGN_9)->FillBindingParameter(MixerparameterAction::TOGGLE, CHANNEL_MUTE, 0);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_10)->FillBindingParameter(MixerparameterAction::TOGGLE, CHANNEL_MUTE, 1);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_11)->FillBindingParameter(MixerparameterAction::TOGGLE, CHANNEL_MUTE, 2);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_12)->FillBindingParameter(MixerparameterAction::TOGGLE, CHANNEL_MUTE, 3);

        bank = assingBanks[(uint)X32AssignBankId::Bank_C];

        bank->bindingMap->at(SurfaceElementId::ASSIGN_ENCODER_1)->FillBindingParameter(MixerparameterAction::CHANGE_SELECTED_CHANNEL, CHANNEL_GAIN, 0);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_ENCODER_2)->FillBindingParameter(MixerparameterAction::CHANGE, CHANNEL_GAIN, 1);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_ENCODER_3)->FillBindingParameter(MixerparameterAction::CHANGE, CHANNEL_GAIN, 2);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_ENCODER_4)->FillBindingParameter(MixerparameterAction::CHANGE, CHANNEL_EQ_FREQ1, 0);

		bank->bindingMap->at(SurfaceElementId::ASSIGN_LCD_1)->FillBindingParameter(MixerparameterAction::LCD_Assign, NONE, 0);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_LCD_2)->FillBindingParameter(MixerparameterAction::LCD_Assign, NONE, 1);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_LCD_3)->FillBindingParameter(MixerparameterAction::LCD_Assign, NONE, 2);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_LCD_4)->FillBindingParameter(MixerparameterAction::LCD_Assign, NONE, 3);

		bank->bindingMap->at(SurfaceElementId::ASSIGN_5)->FillBindingParameter(MixerparameterAction::TOGGLE, CHANNEL_PHANTOM, 0);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_6)->FillBindingParameter(MixerparameterAction::TOGGLE, CHANNEL_PHANTOM, 1);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_7)->FillBindingParameter(MixerparameterAction::TOGGLE, CHANNEL_PHANTOM, 2);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_8)->FillBindingParameter(MixerparameterAction::TOGGLE, CHANNEL_PHANTOM, 3);

        bank->bindingMap->at(SurfaceElementId::ASSIGN_9)->FillBindingParameter(MixerparameterAction::TOGGLE, CHANNEL_LOWCUT_ENABLE, 0);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_10)->FillBindingParameter(MixerparameterAction::TOGGLE, CHANNEL_LOWCUT_FREQ, 0);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_11)->FillBindingParameter(MixerparameterAction::TOGGLE, CHANNEL_EQ_ENABLE, 0);
		bank->bindingMap->at(SurfaceElementId::ASSIGN_12)->FillBindingParameter(MixerparameterAction::TOGGLE, CHANNEL_EQ_FREQ1, 0);
	}
}

OMCAssignBank* ConfigClient::GetAssignBank(X32AssignBankId id)
{
    return assingBanks[(uint)id];
}

/// @brief Checks, if the value of the bound Mixerparameter has changed
/// @param id The surface element which bound Mixerparameter should be checked
/// @return 
bool ConfigClient::HasBoundParameterChanged(SurfaceElementId id)
{
    SurfaceBindingParameter* binding = GetSurfaceBinding(id);    

    MP_ID parameter_id = ParameterCalcId(binding);
	uint parameter_index = ParameterCalcIndex(binding);

    bool hasChanged = mp_changedlist->contains(parameter_id) && mp_changedlist->at(parameter_id).contains(parameter_index);

    if (ParameterDependsOn(binding) != NONE)
    {
        hasChanged |= HasParameterChanged(ParameterDependsOn(binding));
    }

    return hasChanged;
}