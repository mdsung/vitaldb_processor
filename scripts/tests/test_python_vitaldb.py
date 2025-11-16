#!/usr/bin/env python3
"""Python VitalDB 사용법 테스트"""

import vitaldb

# 버퍼 오류 수정
vitaldb.utils.FMT_TYPE_LEN[7] = ("i", 4)
vitaldb.utils.FMT_TYPE_LEN[8] = ("I", 4)

# 샘플 파일 로드
vital_path = './data_sample/MICUA01_240724_180000.vital'
print(f"Loading: {vital_path}")

vf = vitaldb.VitalFile(vital_path)

print(f"\n=== File Info ===")
print(f"dt_start: {vf.dtstart}")
print(f"dt_end: {vf.dtend}")
print(f"duration: {vf.dtend - vf.dtstart if vf.dtend and vf.dtstart else 'N/A'}")

print(f"\n=== Tracks ({len(vf.trks)}) ===")
for idx, (trk_name, trk) in enumerate(list(vf.trks.items())[:5]):
    print(f"\n{idx+1}. {trk_name}")
    print(f"   type: {trk.type}")
    print(f"   fmt: {trk.fmt}")
    print(f"   unit: {trk.unit}")
    print(f"   srate: {trk.srate}")

    # 데이터 접근 방법 확인
    print(f"   hasattr 'vals': {hasattr(trk, 'vals')}")
    print(f"   hasattr 'times': {hasattr(trk, 'times')}")

    if hasattr(trk, 'vals'):
        print(f"   vals type: {type(trk.vals)}")
        print(f"   vals is None: {trk.vals is None}")
        if trk.vals is not None:
            print(f"   vals length: {len(trk.vals)}")
            if len(trk.vals) > 0:
                print(f"   first val: {trk.vals[0]}")

    if hasattr(trk, 'times'):
        print(f"   times type: {type(trk.times)}")
        print(f"   times is None: {trk.times is None}")
        if trk.times is not None:
            print(f"   times length: {len(trk.times)}")
            if len(trk.times) > 0:
                print(f"   first time: {trk.times[0]}")

    # 다른 접근 방법들 시도
    print(f"   dir(trk): {[x for x in dir(trk) if not x.startswith('_')]}")
