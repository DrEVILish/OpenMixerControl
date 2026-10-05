#pragma once

#include <map>

#include "surface.h"
#include "surface-controller.h"

// Surface controller for the Midas PRO1, PRO2C and PRO2.
//
// The link between the PRO's control computer and its surface boards is not
// decoded yet (see DrEVILish/OpenProSeries, docs/hardware). Until it is, this
// controller keeps what OMC asks of the surface (LEDs, fader positions, meter
// LEDs, LCD keys) in a state table, so the link only has to send it, and the
// PRO models run with the GUI alone.
//
// Boards and indexes are OMC's logical ones, see OMC_BOARD_PRO_* in enum.h and
// the PRO block in X32Config::DefineSurfaceElements().
class SurfaceControllerPro : public SurfaceController
{
    public:

        struct LedState { bool on = false; bool blink = false; };
        struct LcdState { uint8_t color = 0; String text; };

        SurfaceControllerPro(X32BaseParameter* basepar);

        void Reset() override;

        void SetFader(uint8_t boardId, uint8_t index, uint16_t position) override;
        void FaderMoved(uint8_t boardId, uint8_t index, uint16_t value) override;
        void FaderReset() override;
        void SetLed(SurfaceElementId buttonOrLed, bool ledOn, bool blink) override;
        void SetMeterLed(uint8_t boardId, uint8_t index, uint8_t leds) override;
        void SetLcd(LcdData* p_data, uint p_textCount) override;

        LedState GetLed(SurfaceElementId buttonOrLed);
        uint16_t GetFader(uint8_t boardId, uint8_t index);
        uint8_t GetMeterLed(uint8_t boardId, uint8_t index);
        LcdState GetLcd(uint8_t boardId, uint8_t index);

    private:

        static uint16_t Key(uint8_t boardId, uint8_t index) { return (boardId << 8) | index; }

        std::map<SurfaceElementId, LedState> leds;
        std::map<uint16_t, uint16_t> faders;
        std::map<uint16_t, uint8_t> meters;
        std::map<uint16_t, LcdState> lcds;
};
