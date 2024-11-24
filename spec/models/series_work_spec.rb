# typed: false
# frozen_string_literal: true

describe SeriesWork, type: :model do
  describe '.sort_season' do
    context 'when same season_year and different started_on' do
      before do
        series = create(:series)

        common_attributes = { season_year: 2010, season_name: 'spring' }
        @work_started_on_first = create(:work,
                                        **common_attributes,
                                        started_on: Date.parse('2010-01-01'))
        @work_started_on_last = create(:work,
                                       **common_attributes,
                                       started_on: Date.parse('2010-12-01'))
        @work_started_on_second = create(:work,
                                         **common_attributes,
                                         started_on: Date.parse('2010-06-01'))

        create(:series_work, series:, work: @work_started_on_first)
        create(:series_work, series:, work: @work_started_on_last)
        create(:series_work, series:, work: @work_started_on_second)
      end

      it 'sorts SeriesWork by started_on column in ascending order', :aggregate_failures do
        expect(SeriesWork.sort_season.first.work).to eq(@work_started_on_first)
        expect(SeriesWork.sort_season.second.work).to eq(@work_started_on_second)
        expect(SeriesWork.sort_season.last.work).to eq(@work_started_on_last)
      end
    end
  end
end
