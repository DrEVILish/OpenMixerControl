#include "surface-controller-pro.h"

SurfaceControllerPro::SurfaceControllerPro(X32BaseParameter* basepar) : SurfaceController(basepar)
{
    helper->DEBUG_SURFACE(DEBUGLEVEL_NORMAL, "SurfaceControllerPro: surface link not implemented yet, running GUI only");
}

void SurfaceControllerPro::Reset()
{
    helper->DEBUG_SURFACE(DEBUGLEVEL_VERBOSE, "SurfaceControllerPro Reset");

    leds.clear();
    faders.clear();
    meters.clear();
    lcds.clear();
}

void SurfaceControllerPro::SetFader(uint8_t boardId, uint8_t index, uint16_t position)
{
    helper->DEBUG_SURFACE(DEBUGLEVEL_TRACE, "SurfaceControllerPro SetFader: board=0x%02X index=%u position=%u", boardId, index, position);
    faders[Key(boardId, index)] = position;
}

void SurfaceControllerPro::FaderMoved(uint8_t boardId, uint8_t index, uint16_t value)
{
    // the fader is where the user put it, so the link must not drive it back
    faders[Key(boardId, index)] = value;
}

void SurfaceControllerPro::FaderReset()
{
    faders.clear();
}

void SurfaceControllerPro::SetLed(SurfaceElementId buttonOrLed, bool ledOn, bool blink)
{
    helper->DEBUG_SURFACE(DEBUGLEVEL_TRACE, "SurfaceControllerPro SetLed: %s on=%d blink=%d", config->GetSurfaceElement(buttonOrLed)->GetName().c_str(), ledOn, blink);
    leds[buttonOrLed] = { ledOn, blink };
}

void SurfaceControllerPro::SetMeterLed(uint8_t boardId, uint8_t index, uint8_t leds)
{
    meters[Key(boardId, index)] = leds;
}

void SurfaceControllerPro::SetLcd(LcdData* p_data, uint p_textCount)
{
    LcdState lcd;
    lcd.color = p_data->color;
    for (uint i = 0; i < p_textCount && i < 7; i++)
    {
        if (p_data->texts[i].text.length() == 0)
        {
            continue;
        }
        if (lcd.text.length() > 0)
        {
            lcd.text += " ";
        }
        lcd.text += p_data->texts[i].text;
    }

    helper->DEBUG_SURFACE(DEBUGLEVEL_TRACE, "SurfaceControllerPro SetLcd: board=0x%02X index=%u color=0x%X text=\"%s\"", p_data->boardId, p_data->lcdIndex, lcd.color, lcd.text.c_str());
    lcds[Key(p_data->boardId, p_data->lcdIndex)] = lcd;
}

SurfaceControllerPro::LedState SurfaceControllerPro::GetLed(SurfaceElementId buttonOrLed)
{
    auto it = leds.find(buttonOrLed);
    return it == leds.end() ? LedState() : it->second;
}

uint16_t SurfaceControllerPro::GetFader(uint8_t boardId, uint8_t index)
{
    auto it = faders.find(Key(boardId, index));
    return it == faders.end() ? 0 : it->second;
}

uint8_t SurfaceControllerPro::GetMeterLed(uint8_t boardId, uint8_t index)
{
    auto it = meters.find(Key(boardId, index));
    return it == meters.end() ? 0 : it->second;
}

SurfaceControllerPro::LcdState SurfaceControllerPro::GetLcd(uint8_t boardId, uint8_t index)
{
    auto it = lcds.find(Key(boardId, index));
    return it == lcds.end() ? LcdState() : it->second;
}

// ######## ########  ######  ########  ######  
//    ##    ##       ##    ##    ##    ##    ## 
//    ##    ##       ##          ##    ##       
//    ##    ######    ######     ##     ######  
//    ##    ##             ##    ##          ## 
//    ##    ##       ##    ##    ##    ##    ## 
//    ##    ########  ######     ##     ######  

#include "../lib_ext/doctest/doctest/doctest.h"

TEST_CASE("PRO surface elements")
{
    Helper* helper = new Helper();

    SUBCASE("PRO1 has two bays of 8 strips, banked like an X32 Compact")
    {
        X32Config* config = new X32Config("PRO1", helper);

        CHECK(config->HasXM32StyleSurface());
        CHECK(config->HasSurface8InputStrips());
        CHECK_FALSE(config->HasSurface16InputStrips());

        CHECK(config->GetSurfaceElementFader(OMC_BOARD_PRO_INPUT, 0)->GetId() == SurfaceElementId::BOARD_L_FADER_1);
        CHECK(config->GetSurfaceElementFader(OMC_BOARD_PRO_OUTPUT, 7)->GetId() == SurfaceElementId::BOARD_R_FADER_8);
        CHECK(config->GetSurfaceElementFader(OMC_BOARD_PRO_INPUT2, 0) == nullptr);
        CHECK(config->GetSurfaceElementButton_Wing(OMC_BOARD_PRO_INPUT, 0x10)->GetId() == SurfaceElementId::BOARD_L_MUTE_1);
        CHECK(config->GetSurfaceElementButton_Wing(OMC_BOARD_PRO_OUTPUT, 0x20)->GetId() == SurfaceElementId::PRO_MIX_VCA);
        CHECK(config->GetSurfaceElementButton_Wing(OMC_BOARD_PRO_CENTRE, 0x05)->GetId() == SurfaceElementId::PRO_SCREEN_6);
        CHECK(config->GetSurfaceElementEncoder(OMC_BOARD_PRO_INPUT, 7)->GetId() == SurfaceElementId::PRO_ENCODER_8);
    }

    SUBCASE("PRO2 adds a second input bay")
    {
        X32Config* config = new X32Config("PRO2", helper);

        CHECK(config->HasSurface16InputStrips());
        CHECK(config->GetSurfaceElementFader(OMC_BOARD_PRO_INPUT2, 3)->GetId() == SurfaceElementId::BOARD_M_FADER_4);
    }

    SUBCASE("X32 models have no PRO elements")
    {
        X32Config* config = new X32Config("X32C", helper);

        CHECK_FALSE(config->IsModelAnyPro());
        CHECK(config->HasXM32StyleSurface());
        CHECK(config->GetSurfaceElementButton_Wing(OMC_BOARD_PRO_OUTPUT, 0x20) == nullptr);
    }
}

TEST_CASE("SurfaceControllerPro state")
{
    State* state = new State();
    Helper* helper = new Helper();
    X32Config* config = new X32Config("PRO1", helper);
    X32BaseParameter* basepar = new X32BaseParameter(nullptr, config, state, helper);
    SurfaceControllerPro* pro = new SurfaceControllerPro(basepar);

    pro->SetLed(SurfaceElementId::BOARD_L_MUTE_1, true, false);
    pro->SetLed(SurfaceElementId::PRO_POP_1, true, true);
    pro->SetFader(OMC_BOARD_PRO_INPUT, 2, 3000);
    pro->SetMeterLed(OMC_BOARD_PRO_OUTPUT, 1, 0x3F);

    LcdData lcd;
    lcd.boardId = OMC_BOARD_PRO_INPUT;
    lcd.lcdIndex = 0;
    lcd.color = 0x03;
    lcd.texts[0].text = "Ch01";
    lcd.texts[1].text = "Kick";
    pro->SetLcd(&lcd, 2);

    CHECK(pro->GetLed(SurfaceElementId::BOARD_L_MUTE_1).on);
    CHECK_FALSE(pro->GetLed(SurfaceElementId::BOARD_L_MUTE_1).blink);
    CHECK(pro->GetLed(SurfaceElementId::PRO_POP_1).blink);
    CHECK_FALSE(pro->GetLed(SurfaceElementId::BOARD_L_SOLO_1).on);
    CHECK(pro->GetFader(OMC_BOARD_PRO_INPUT, 2) == 3000);
    CHECK(pro->GetMeterLed(OMC_BOARD_PRO_OUTPUT, 1) == 0x3F);
    CHECK(pro->GetLcd(OMC_BOARD_PRO_INPUT, 0).text == "Ch01 Kick");
    CHECK(pro->GetLcd(OMC_BOARD_PRO_INPUT, 0).color == 0x03);

    SUBCASE("a moved fader keeps where the user put it")
    {
        pro->FaderMoved(OMC_BOARD_PRO_INPUT, 2, 1234);
        CHECK(pro->GetFader(OMC_BOARD_PRO_INPUT, 2) == 1234);
    }

    SUBCASE("reset clears the state")
    {
        pro->Reset();
        CHECK_FALSE(pro->GetLed(SurfaceElementId::BOARD_L_MUTE_1).on);
        CHECK(pro->GetFader(OMC_BOARD_PRO_INPUT, 2) == 0);
        CHECK(pro->GetLcd(OMC_BOARD_PRO_INPUT, 0).text == "");
    }
}
