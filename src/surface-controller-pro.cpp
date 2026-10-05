#include "surface-controller-pro.h"

SurfaceControllerPro::SurfaceControllerPro(X32BaseParameter* basepar) : SurfaceController(basepar)
{
    helper->DEBUG_SURFACE(DEBUGLEVEL_NORMAL, "SurfaceControllerPro: surface link not implemented yet, running GUI only");
}

void SurfaceControllerPro::Reset()
{
    helper->DEBUG_SURFACE(DEBUGLEVEL_VERBOSE, "SurfaceControllerPro Reset");
}

void SurfaceControllerPro::SetFader(uint8_t boardId, uint8_t index, uint16_t position)
{
    helper->DEBUG_SURFACE(DEBUGLEVEL_TRACE, "SurfaceControllerPro SetFader: board=%u index=%u position=%u", boardId, index, position);
}

void SurfaceControllerPro::SetLed(SurfaceElementId buttonOrLed, bool ledOn, bool blink)
{
    helper->DEBUG_SURFACE(DEBUGLEVEL_TRACE, "SurfaceControllerPro SetLed: element=%u on=%d blink=%d", (uint)buttonOrLed, ledOn, blink);
}

void SurfaceControllerPro::SetMeterLed(uint8_t boardId, uint8_t index, uint8_t leds)
{
}

void SurfaceControllerPro::SetLcd(LcdData* p_data, uint p_textCount)
{
    helper->DEBUG_SURFACE(DEBUGLEVEL_TRACE, "SurfaceControllerPro SetLcd: texts=%u", p_textCount);
}
