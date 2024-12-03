# typed: false
# frozen_string_literal: true

def create_work(season_year: 2010, season_name: "spring", started_on: "2010-04-11")
  create(:work,
    season_year:,
    season_name:,
    started_on: Date.parse(started_on))
end

describe SeriesWork, type: :model do
  describe ".sort_season" do
    let(:series) { create(:series) }

    context "differ in season_year, season_name, and started_on" do
      let(:work_earier_started_on) { create_work(started_on: "2010-04-01") }
      let(:work_later_started_on) { create_work(started_on: "2010-04-21") }
      let(:work_earier_season_year) { create_work(season_year: 2009) }
      let(:work_earier_season_name) { create_work(season_name: "winter", started_on: "2010-01-21") }

      before do
        create(:series_work, series:, work: work_later_started_on)
        create(:series_work, series:, work: work_earier_started_on)
        create(:series_work, series:, work: work_earier_season_year)
        create(:series_work, series:, work: work_earier_season_name)
      end

      it "sorts SeriesWork by season_year, then season_name, and then started_on in ascending order",
        :aggregate_failures do
        expect(SeriesWork.sort_season[0].work).to eq(work_earier_season_year)
        expect(SeriesWork.sort_season[1].work).to eq(work_earier_season_name)
        expect(SeriesWork.sort_season[2].work).to eq(work_earier_started_on)
        expect(SeriesWork.sort_season[3].work).to eq(work_later_started_on)
      end
    end

    context "when one work has a nil season_year but an earlier started_on date" do
      let(:work_with_season_year_nil) { create_work(season_year: nil, started_on: "2009-04-01") }
      let(:work_with_season_year_exists) { create_work(season_year: 2010, started_on: "2010-04-01") }

      before do
        create(:series_work, series:, work: work_with_season_year_nil)
        create(:series_work, series:, work: work_with_season_year_exists)
      end

      it "places the work with a valid season_year before the work with a nil season_year, regardless of started_on date",
        :aggregate_failures do
        expect(SeriesWork.sort_season[0].work).to eq(work_with_season_year_exists)
        expect(SeriesWork.sort_season[1].work).to eq(work_with_season_year_nil)
      end
    end
  end
end
