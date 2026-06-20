#include "config-server.h"

//#####################################################################################################################
//
// ##        #######     ###    ########  
// ##       ##     ##   ## ##   ##     ## 
// ##       ##     ##  ##   ##  ##     ## 
// ##       ##     ## ##     ## ##     ## 
// ##       ##     ## ######### ##     ## 
// ##       ##     ## ##     ## ##     ## 
// ########  #######  ##     ## ########  
//
//#####################################################################################################################

bool ConfigServer::LoadConfig(uint scene)
{
    WString::String loadFile = String("scene") + String(scene) + String(".json");

    // no file found
	if (helper->GetFileSize(loadFile.c_str()) == -1)
    {
        return false;
    }

    if (access(loadFile.c_str(), F_OK) == -1)
    {
        helper->Error("Can not load X32Config. File %s does not exist.", loadFile.c_str());
    }

    // Read file
    ifstream ifs(loadFile.c_str());
    ostringstream oss;
    oss << ifs.rdbuf();
    std::string entireFile = oss.str();
    ifs.close();  

    // Parse JSON
    vector<X32ConfigFileEntry> entries;
    auto error = glz::read_json(entries, entireFile);
    if (error) {
       std::string error_msg = glz::format_error(error, entireFile);
       std::cout << "ERROR: " << error_msg << std::endl;
    }
    
    // Fill Mixerparameter
	for (uint i=0; i < entries.size(); i++)
	{
		MP_ID parameter_id = entries.at(i).MixerparameterId;    
		Mixerparameter* parameter = GetParameter(parameter_id);

        switch(parameter->GetType())
        {
            case MP_VALUE_TYPE::STRING:
                parameter->Config_SetValueString(entries.at(i).string_value);
                break;
                
            default:
                parameter->Config_SetValue(entries.at(i).value);
        }

        Refresh(parameter_id, 0);
	}

    return true;
}

//#####################################################################################################################
//
//  ######     ###    ##     ## ######## 
// ##    ##   ## ##   ##     ## ##       
// ##        ##   ##  ##     ## ##       
//  ######  ##     ## ##     ## ######   
//       ## #########  ##   ##  ##       
// ##    ## ##     ##   ## ##   ##       
//  ######  ##     ##    ###    ######## 
//
//#####################################################################################################################

void ConfigServer::Save(uint scene)
{
    String saveFile = String("scene") + String(scene) + String(".json");
    helper->DEBUG_INI(DEBUGLEVEL_NORMAL, "Save config to %s", saveFile.c_str());

    vector<X32ConfigFileEntry*> entries;

	// go over all known Mixerparameter an store them
	for (uint i=0; i < (uint)__ELEMENT_COUNTER_DO_NOT_MOVE; i++)
	{
        Mixerparameter* parameter = GetParameter((MP_ID)i);

        if (parameter->GetId() == NONE || parameter->IsNoConfig())
        {
            // this Mixerparameter should not be written to config file
            continue;
        }

        X32ConfigFileEntry* entry = new X32ConfigFileEntry();
        entry->MixerparameterId = (MP_ID)i;
        //entry->key = parameter->GetConfigEntry().c_str();
        entry->MixerparameterName = parameter->GetName();

        switch(parameter->GetType())
        {
            case MP_VALUE_TYPE::STRING:
                entry->string_value = parameter->Config_GetValueString();
                break;
            default:
                entry->value = parameter->Config_GetValue();
        }

        entries.push_back(entry);
	}

    std::string json;
    //auto error = glz::write<glz::opts{.prettify = true}>(entries, json);
    auto error = glz::write_json(entries, json);
    if (error) {
       std::string error_msg = glz::format_error(error, json);
       std::cout << error_msg << std::endl;
    }

    std::ofstream out(saveFile.c_str());
    out << json;
    out.close();
}