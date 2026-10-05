#pragma once

#include "surface.h"
#include "surface-controller.h"

// Surface controller for the Midas PRO1, PRO2C and PRO2.
//
// The link between the PRO's control computer and its surface boards is not
// decoded yet (see DrEVILish/OpenProSeries, docs/hardware). Until it is, this
// controller only logs what OMC asks of the surface, so the PRO models run
// with the GUI alone.
class SurfaceControllerPro : public SurfaceController
{
    public:

        SurfaceControllerPro(X32BaseParameter* basepar);

        void Reset() override;

        void SetFader(uint8_t boardId, uint8_t index, uint16_t position) override;
        void SetLed(SurfaceElementId buttonOrLed, bool ledOn, bool blink) override;
        void SetMeterLed(uint8_t boardId, uint8_t index, uint8_t leds) override;
        void SetLcd(LcdData* p_data, uint p_textCount) override;
};
