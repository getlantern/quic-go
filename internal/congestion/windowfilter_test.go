package congestion

import (
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func updateWithIrrelevantSamples(wF *windowFilter, maxValue int64, t time.Time) {
	for i := int64(0); i < 1000; i++ {
		wF.Update(i%maxValue, t)
	}
}

func mkMinFilter() windowFilter {
	wF := windowFilter{
		MaxOrMinFilter: false,
		windowLength:   time.Millisecond * 99,
	}
	testTime := time.Date(2021, time.February, 1, 10, 10, 10, 10, time.Now().UTC().Location())
	rttSample := time.Duration(time.Millisecond * 10)
	for i := 0; i < 5; i++ {
		wF.Update(int64(rttSample), testTime)
		testTime = testTime.Add(time.Millisecond * 25)
		rttSample += time.Duration(time.Millisecond * 10)
	}
	return wF
}

func mkMaxFilter() windowFilter {
	wF := windowFilter{
		MaxOrMinFilter: true,
		windowLength:   time.Millisecond * 99,
	}
	testTime := time.Date(2021, time.February, 1, 10, 10, 10, 10, time.Now().UTC().Location())
	bps := 1000

	for i := 0; i < 5; i++ {
		wF.Update(int64(bps), testTime)

		testTime = testTime.Add(time.Millisecond * 25)
		bps -= 100
	}

	if wF.GetBest() != 900 {
		panic("wF.GetBest() != 900")
	}
	if wF.GetSecondBest() != 700 {
		panic("wF.GetSecondBest() != 700")
	}
	if wF.GetThirdBest() != 600 {
		panic("wF.GetThirdBest() != 600")
	}
	return wF
}

var _ = Describe("WindowFilter", func() {
	It("Has Zerod Default Values", func() {
		wF := windowFilter{}

		Expect(wF.GetBest()).To(Equal(int64(0)))
		Expect(wF.GetSecondBest()).To(Equal(int64(0)))
		Expect(wF.GetThirdBest()).To(Equal(int64(0)))
	})

	It("MonotonicallyIncreasingMin", func() {
		wF := windowFilter{
			MaxOrMinFilter: false,
			windowLength:   time.Millisecond * 99,
		}
		testTime := time.Now()
		rttSample := time.Duration(time.Millisecond * 10)
		wF.Update(int64(rttSample), testTime)
		Expect(wF.GetBest()).To(Equal(int64(rttSample)))
		os.Stderr.WriteString(fmt.Sprintf("\nI: %d\tSample: %d\t \n mins:\n%d\t%d\t%d\n", -1, int64(rttSample), wF.GetBest(), wF.GetSecondBest(), wF.GetThirdBest()))

		// Gradually increase the rtt samples and ensure the windowed min rtt starts
		// rising.
		for i := 0; i < 6; i++ {
			testTime = testTime.Add(time.Millisecond * 25)
			rttSample += time.Duration(time.Millisecond * 10)
			wF.Update(int64(rttSample), testTime)

			os.Stderr.WriteString(fmt.Sprintf("\nI: %d\tSample: %d\t \n mins:\n%d\t%d\t%d\n", i, int64(rttSample), wF.GetBest(), wF.GetSecondBest(), wF.GetThirdBest()))

			if i < 3 {
				Expect(wF.GetBest()).To(Equal(int64(time.Duration(time.Millisecond * 10))))
			} else if i == 3 {
				Expect(wF.GetBest()).To(Equal(int64(time.Duration(time.Millisecond * 20))))
			} else if i < 6 {
				Expect(wF.GetBest()).To(Equal(int64(time.Duration(time.Millisecond * 40))))
			}
		}
	})

	It("MonotonicallyDecreasingMax", func() {
		wF := windowFilter{
			MaxOrMinFilter: true,
			windowLength:   time.Millisecond * 99,
		}
		testTime := time.Now()
		bps := 1000
		wF.Update(int64(bps), testTime)
		Expect(wF.GetBest()).To(Equal(int64(bps)))

		os.Stderr.WriteString(fmt.Sprintf("\nI: %d\tSample: %d\t \n mins:\n%d\t%d\t%d\n", -1, int64(bps), wF.GetBest(), wF.GetSecondBest(), wF.GetThirdBest()))

		// Gradually decrease the bw samples and ensure the windowed max bw starts
		// decreasing.

		for i := 0; i < 6; i++ {
			testTime = testTime.Add(time.Millisecond * 25)
			bps = bps - 100
			wF.Update(int64(bps), testTime)

			os.Stderr.WriteString(fmt.Sprintf("\nI: %d\tSample: %d\t \n mins:\n%d\t%d\t%d\n", i, int64(bps), wF.GetBest(), wF.GetSecondBest(), wF.GetThirdBest()))

			if i < 3 {
				Expect(wF.GetBest()).To(Equal(int64(1000)))
			} else if i == 3 {
				Expect(wF.GetBest()).To(Equal(int64(900)))
			} else if i < 6 {
				Expect(wF.GetBest()).To(Equal(int64(700)))
			}
		}
	})

	It("SampleChangesThirdBestMin", func() {
		wF := mkMinFilter()
		rttSample := wF.GetThirdBest() - int64(time.Millisecond*5)
		testTime := time.Date(2021, time.February, 1, 10, 10, 10, 10, time.Now().UTC().Location())
		testTime = testTime.Add(time.Millisecond * 101)
		wF.Update(rttSample, testTime)

		Expect(wF.GetThirdBest()).To(Equal(int64(rttSample)))
		Expect(wF.GetSecondBest()).To(Equal(int64(time.Millisecond * 40)))
		Expect(wF.GetBest()).To(Equal(int64(time.Millisecond * 20)))
	})

	It("SampleChangesThirdBestMax", func() {
		wF := mkMaxFilter()
		bps := wF.GetThirdBest() + 50
		testTime := time.Date(2021, time.February, 1, 10, 10, 10, 10, time.Now().UTC().Location())
		testTime = testTime.Add(time.Millisecond * 101)
		wF.Update(bps, testTime)

		Expect(wF.GetThirdBest()).To(Equal(int64(bps)))
		Expect(wF.GetSecondBest()).To(Equal(int64(700)))
		Expect(wF.GetBest()).To(Equal(int64(900)))
	})

	It("SampleChangesSecondBestMin", func() {
		wF := mkMinFilter()
		rttSample := wF.GetSecondBest() - int64(time.Millisecond*5)
		testTime := time.Date(2021, time.February, 1, 10, 10, 10, 10, time.Now().UTC().Location())
		testTime = testTime.Add(time.Millisecond * 101)
		wF.Update(rttSample, testTime)

		Expect(wF.GetThirdBest()).To(Equal(int64(rttSample)))
		Expect(wF.GetSecondBest()).To(Equal(int64(rttSample)))
		Expect(wF.GetBest()).To(Equal(int64(time.Millisecond * 20)))
	})

	It("SampleChangesSecondBestMax", func() {
		wF := mkMaxFilter()
		bps := wF.GetSecondBest() + 50
		testTime := time.Date(2021, time.February, 1, 10, 10, 10, 10, time.Now().UTC().Location())
		testTime = testTime.Add(time.Millisecond * 101)
		wF.Update(bps, testTime)

		Expect(wF.GetThirdBest()).To(Equal(int64(bps)))
		Expect(wF.GetSecondBest()).To(Equal(int64(bps)))
		Expect(wF.GetBest()).To(Equal(int64(900)))
	})

	It("SampleChangesAllMins", func() {
		wF := mkMinFilter()
		rttSample := wF.GetBest() - int64(time.Millisecond*5)
		testTime := time.Date(2021, time.February, 1, 10, 10, 10, 10, time.Now().UTC().Location())
		testTime = testTime.Add(time.Millisecond * 101)
		wF.Update(rttSample, testTime)

		Expect(wF.GetThirdBest()).To(Equal(int64(rttSample)))
		Expect(wF.GetSecondBest()).To(Equal(int64(rttSample)))
		Expect(wF.GetBest()).To(Equal(int64(rttSample)))
	})

	It("SampleChangesAllMaxs", func() {
		wF := mkMaxFilter()
		bps := wF.GetBest() + 50
		testTime := time.Date(2021, time.February, 1, 10, 10, 10, 10, time.Now().UTC().Location())
		testTime = testTime.Add(time.Millisecond * 101)
		wF.Update(bps, testTime)

		Expect(wF.GetThirdBest()).To(Equal(int64(bps)))
		Expect(wF.GetSecondBest()).To(Equal(int64(bps)))
		Expect(wF.GetBest()).To(Equal(int64(bps)))
	})

	It("ExpireBestMin", func() {
		wF := mkMinFilter()
		old3rdBest := wF.GetThirdBest()
		old2ndBest := wF.GetSecondBest()
		rttSample := old3rdBest + int64(time.Millisecond*5)
		testTime := time.Date(2021, time.February, 1, 10, 10, 10, 10, time.Now().UTC().Location())
		testTime = testTime.Add(time.Millisecond * 125)
		wF.Update(rttSample, testTime)

		Expect(wF.GetThirdBest()).To(Equal(int64(rttSample)))
		Expect(wF.GetSecondBest()).To(Equal(int64(old3rdBest)))
		Expect(wF.GetBest()).To(Equal(int64(old2ndBest)))
	})

	It("ExpireBestMax", func() {
		wF := mkMaxFilter()
		old3rdBest := wF.GetThirdBest()
		old2ndBest := wF.GetSecondBest()
		bps := old3rdBest - 50
		testTime := time.Date(2021, time.February, 1, 10, 10, 10, 10, time.Now().UTC().Location())
		testTime = testTime.Add(time.Millisecond * 125)
		wF.Update(bps, testTime)

		Expect(wF.GetThirdBest()).To(Equal(int64(bps)))
		Expect(wF.GetSecondBest()).To(Equal(int64(old3rdBest)))
		Expect(wF.GetBest()).To(Equal(int64(old2ndBest)))
	})

	It("ExpireSecondBestMin", func() {
		wF := mkMinFilter()
		old3rdBest := wF.GetThirdBest()
		rttSample := old3rdBest + int64(time.Millisecond*5)
		testTime := time.Date(2021, time.February, 1, 10, 10, 10, 10, time.Now().UTC().Location())
		testTime = testTime.Add(time.Millisecond * 175)
		wF.Update(rttSample, testTime)

		Expect(wF.GetThirdBest()).To(Equal(int64(rttSample)))
		Expect(wF.GetSecondBest()).To(Equal(int64(rttSample)))
		Expect(wF.GetBest()).To(Equal(int64(old3rdBest)))
	})

	It("ExpireSecondBestMax", func() {
		wF := mkMaxFilter()
		old3rdBest := wF.GetThirdBest()
		bps := old3rdBest - 50
		testTime := time.Date(2021, time.February, 1, 10, 10, 10, 10, time.Now().UTC().Location())
		testTime = testTime.Add(time.Millisecond * 175)
		wF.Update(bps, testTime)

		Expect(wF.GetThirdBest()).To(Equal(int64(bps)))
		Expect(wF.GetSecondBest()).To(Equal(int64(bps)))
		Expect(wF.GetBest()).To(Equal(int64(old3rdBest)))
	})

	It("ExpireAllMins", func() {
		wF := mkMinFilter()
		rttSample := wF.GetThirdBest() + int64(time.Millisecond*5)

		testTime := time.Date(2021, time.February, 1, 10, 10, 10, 10, time.Now().UTC().Location())
		testTime = testTime.Add(time.Millisecond * 200)
		wF.Update(rttSample, testTime)

		Expect(wF.GetThirdBest()).To(Equal(int64(rttSample)))
		Expect(wF.GetSecondBest()).To(Equal(int64(rttSample)))
		Expect(wF.GetBest()).To(Equal(int64(rttSample)))
	})

	It("ExpireAllMaxs", func() {
		wF := mkMaxFilter()
		bps := wF.GetThirdBest() + 50

		testTime := time.Date(2021, time.February, 1, 10, 10, 10, 10, time.Now().UTC().Location())
		testTime = testTime.Add(time.Millisecond * 200)
		wF.Update(bps, testTime)

		Expect(wF.GetThirdBest()).To(Equal(int64(bps)))
		Expect(wF.GetSecondBest()).To(Equal(int64(bps)))
		Expect(wF.GetBest()).To(Equal(int64(bps)))
	})

	It("ExpireCounterBasedMax", func() {
		Best := 50000
		testTime := time.Date(2021, time.February, 1, 10, 10, 10, 10, time.Now().UTC().Location())

		wF := windowFilter{
			MaxOrMinFilter: true,
			windowLength:   time.Millisecond * 2,
		}

		// Insert 50000 at t = 1.
		testTime = testTime.Add(time.Millisecond * 1)
		wF.Update(50000, testTime)
		Expect(wF.GetBest()).To(Equal(int64(Best)))
		updateWithIrrelevantSamples(&wF, 20, testTime)
		Expect(wF.GetBest()).To(Equal(int64(Best)))

		// Insert 40000 at t = 2.  Nothing is expected to expire.
		testTime = time.Date(2021, time.February, 1, 10, 10, 10, 10, time.Now().UTC().Location())
		testTime = testTime.Add(time.Millisecond * 2)
		wF.Update(40000, testTime)
		Expect(wF.GetBest()).To(Equal(int64(Best)))
		updateWithIrrelevantSamples(&wF, 20, testTime)
		Expect(wF.GetBest()).To(Equal(int64(Best)))

		// Insert 30000 at t = 3.  Nothing is expected to expire yet.
		testTime = time.Date(2021, time.February, 1, 10, 10, 10, 10, time.Now().UTC().Location())
		testTime = testTime.Add(time.Millisecond * 3)
		wF.Update(30000, testTime)
		Expect(wF.GetBest()).To(Equal(int64(Best)))
		updateWithIrrelevantSamples(&wF, 20, testTime)
		Expect(wF.GetBest()).To(Equal(int64(Best)))

		testTime = time.Date(2021, time.February, 1, 10, 10, 10, 10, time.Now().UTC().Location())
		testTime = testTime.Add(time.Millisecond * 4)
		NewBest := 40000
		wF.Update(20000, testTime)
		Expect(wF.GetBest()).To(Equal(int64(NewBest)))
		updateWithIrrelevantSamples(&wF, 20, testTime)
		Expect(wF.GetBest()).To(Equal(int64(NewBest)))
	})
})
