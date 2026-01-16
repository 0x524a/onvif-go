# Test Reports

This directory contains test reports generated from real camera testing.

## Files

- **camera_test_report_Bosch_FLEXIDOME_indoor_5100i_IR_20251201_234919.json** - Initial test report
- **camera_test_report_Bosch_FLEXIDOME_indoor_5100i_IR_20251201_235612.json** - Extended test report
- **camera_test_report_Bosch_FLEXIDOME_indoor_5100i_IR_20251202_000918.json** - Comprehensive test report

## Camera Information

**Manufacturer:** Bosch  
**Model:** FLEXIDOME indoor 5100i IR  
**Firmware Version:** 8.71.0066  
**Serial Number:** 404754734001050102  
**Hardware ID:** F000B543  
**IP Address:** 192.168.1.201

## Report Format

Each JSON report contains:
- Device information (manufacturer, model, firmware, etc.)
- Test results for all operations tested
- Success/failure status for each operation
- Response times
- Error messages (if any)
- Parsed response data

## Generating Reports

To generate new test reports, run:

```bash
go run examples/test-real-camera-all/main.go
```

Reports are automatically saved with timestamps in the filename.

---

*Last Updated: December 2, 2025*

