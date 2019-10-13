# frozen_string_literal: true

describe EpisodeGeneratorService, type: :service do
  context "New work" do
    let(:channel) { create(:channel) }
    let(:work) { create(:work) }

    context "has no episodes" do
      let(:program_detail) { create(:program_detail, channel: channel, work: work, started_at: Time.parse("2019-01-04 0:00:00")) }
      let!(:program1) { create(:program, program_detail: program_detail, channel: channel, work: work, episode: nil, number: 1, started_at: Time.parse("2019-01-04 0:00:00")) }
      let!(:program2) { create(:program, program_detail: program_detail, channel: channel, work: work, episode: nil, number: 2, started_at: Time.parse("2019-01-11 0:00:00")) }

      context "when run on 2019-01-01" do
        before do
          EpisodeGeneratorService.execute!(now: Time.parse("2019-01-01 0:00:00"))
        end

        it "creates 1 episode" do
          episodes = work.episodes.published.order(:raw_number)
          expect(episodes.count).to eq(1)
          expect(episodes[0].raw_number).to eq(1.0)
        end
      end

      context "when run on 2019-01-08" do
        before do
          EpisodeGeneratorService.execute!(now: Time.parse("2019-01-08 0:00:00"))
        end

        it "creates 2 episodes" do
          episodes = work.episodes.published.order(:raw_number)
          expect(episodes.count).to eq(2)
          expect(episodes[0].raw_number).to eq(1.0)
          expect(episodes[1].raw_number).to eq(2.0)
        end
      end
    end

    context "has episodes but no irregular" do
      let(:program_detail) { create(:program_detail, channel: channel, work: work, started_at: Time.parse("2019-01-04 0:00:00")) }
      let(:episode1) { create(:episode, work: work, raw_number: 1.0) }
      let!(:program1) { create(:program, program_detail: program_detail, channel: channel, work: work, episode: episode1, number: 1, started_at: Time.parse("2019-01-04 0:00:00")) }
      let!(:program2) { create(:program, program_detail: program_detail, channel: channel, work: work, episode: nil, number: 2, started_at: Time.parse("2019-01-11 0:00:00")) }

      context "when run on 2019-01-01" do
        before do
          EpisodeGeneratorService.execute!(now: Time.parse("2019-01-01 0:00:00"))
        end

        it "does not create episodes" do
          episodes = work.episodes.published.order(:raw_number)
          expect(episodes.count).to eq(1)
          expect(episodes[0].raw_number).to eq(1.0)
        end
      end

      context "when run on 2019-01-08" do
        before do
          EpisodeGeneratorService.execute!(now: Time.parse("2019-01-08 0:00:00"))
        end

        it "creates 1 episode" do
          episodes = work.episodes.published.order(:raw_number)
          expect(episodes.count).to eq(2)
          expect(episodes[0].raw_number).to eq(1.0)
          expect(episodes[1].raw_number).to eq(2.0)
        end
      end
    end

    context "has irregular episodes" do
      let(:program_detail) { create(:program_detail, channel: channel, work: work, started_at: Time.parse("2019-01-04 0:00:00")) }
      let(:episode1) { create(:episode, work: work, raw_number: 1.0,) }
      let(:episode2) { create(:episode, work: work, raw_number: 1.5, title: "2話目から総集編！") }
      let!(:program1) { create(:program, program_detail: program_detail, channel: channel, work: work, episode: episode1, number: 1, started_at: Time.parse("2019-01-04 0:00:00")) }
      let!(:program2) { create(:program, program_detail: program_detail, channel: channel, work: work, episode: episode2, number: 2, started_at: Time.parse("2019-01-11 0:00:00"), irregular: true) }
      let!(:program3) { create(:program, program_detail: program_detail, channel: channel, work: work, episode: nil, number: 3, started_at: Time.parse("2019-01-18 0:00:00")) }

      context "when run on 2019-01-08" do
        before do
          EpisodeGeneratorService.execute!(now: Time.parse("2019-01-08 0:00:00"))
        end

        it "does not create episodes" do
          episodes = work.episodes.published.order(:raw_number)
          expect(episodes.count).to eq(2)
          expect(episodes[0].raw_number).to eq(1.0)
          expect(episodes[1].raw_number).to eq(1.5)
        end
      end

      context "when run on 2019-01-15" do
        before do
          EpisodeGeneratorService.execute!(now: Time.parse("2019-01-15 0:00:00"))
        end

        it "creates 1 episode" do
          episodes = work.episodes.published.order(:raw_number)
          expect(episodes.count).to eq(3)
          expect(episodes[0].raw_number).to eq(1.0)
          expect(episodes[1].raw_number).to eq(1.5)
          expect(episodes[2].raw_number).to eq(2.0)
        end
      end
    end
  end

  context "Old work" do
    let(:channel) { create(:channel) }
    let(:work) { create(:work) }

    context "has no irregular episodes" do
      let(:program_detail) { create(:program_detail, channel: channel, work: work, started_at: Time.parse("2018-04-01 0:00:00"), minimum_episode_generatable_number: 35) }
      let(:episode1) { create(:episode, work: work, raw_number: nil, sort_number: 100) }
      let(:episode2) { create(:episode, work: work, raw_number: nil, sort_number: 200) }
      let!(:episode35) { create(:episode, work: work, raw_number: 35.0, sort_number: 3500) }
      let!(:program1) { create(:program, program_detail: nil, channel: channel, work: work, episode: episode1, number: nil, started_at: Time.parse("2018-04-01 0:00:00")) }
      let!(:program2) { create(:program, program_detail: nil, channel: channel, work: work, episode: episode2, number: nil, started_at: Time.parse("2018-04-08 0:00:00")) }
      let!(:program35) { create(:program, program_detail: program_detail, channel: channel, work: work, episode: nil, number: 35, started_at: Time.parse("2019-01-04 0:00:00")) }
      let!(:program36) { create(:program, program_detail: program_detail, channel: channel, work: work, episode: nil, number: 36, started_at: Time.parse("2019-01-11 0:00:00")) }

      context "when run on 2019-01-01" do
        before do
          EpisodeGeneratorService.execute!(now: Time.parse("2019-01-01 0:00:00"))
        end

        it "does not create episodes" do
          episodes = work.episodes.published.order(:sort_number)
          expect(episodes.count).to eq(3)
          expect(episodes[0].id).to eq(episode1.id)
          expect(episodes[0].raw_number).to eq(nil)
          expect(episodes[1].id).to eq(episode2.id)
          expect(episodes[1].raw_number).to eq(nil)
          expect(episodes[2].id).to eq(episode35.id)
          expect(episodes[2].raw_number).to eq(35.0)
        end
      end

      context "when run on 2019-01-08" do
        before do
          EpisodeGeneratorService.execute!(now: Time.parse("2019-01-08 0:00:00"))
        end

        it "creates 1 episode" do
          episodes = work.episodes.published.order(:sort_number)
          expect(episodes.count).to eq(4)
          expect(episodes[0].id).to eq(episode1.id)
          expect(episodes[0].raw_number).to eq(nil)
          expect(episodes[1].id).to eq(episode2.id)
          expect(episodes[1].raw_number).to eq(nil)
          expect(episodes[2].id).to eq(episode35.id)
          expect(episodes[2].raw_number).to eq(35.0)
          expect(episodes[3].raw_number).to eq(36.0)
        end
      end
    end

    context "has irregular episodes" do
      let(:program_detail) { create(:program_detail, channel: channel, work: work, started_at: Time.parse("2018-04-01 0:00:00"), minimum_episode_generatable_number: 35) }
      let(:episode1) { create(:episode, work: work, raw_number: nil, sort_number: 100) }
      let(:episode2) { create(:episode, work: work, raw_number: nil, sort_number: 200, title: "2話目から総集編！") }
      let!(:episode35) { create(:episode, work: work, raw_number: 35.0, sort_number: 3500) }
      let!(:program1) { create(:program, program_detail: nil, channel: channel, work: work, episode: episode1, number: nil, started_at: Time.parse("2018-04-01 0:00:00")) }
      let!(:program2) { create(:program, program_detail: program_detail, channel: channel, work: work, episode: episode2, number: 2, started_at: Time.parse("2018-04-08 0:00:00"), irregular: true) }
      let!(:program35) { create(:program, program_detail: program_detail, channel: channel, work: work, episode: nil, number: 35, started_at: Time.parse("2019-01-04 0:00:00")) }
      let!(:program36) { create(:program, program_detail: program_detail, channel: channel, work: work, episode: nil, number: 36, started_at: Time.parse("2019-01-11 0:00:00")) }

      context "when run on 2019-01-01" do
        before do
          EpisodeGeneratorService.execute!(now: Time.parse("2019-01-01 0:00:00"))
        end

        it "does not create episodes" do
          episodes = work.episodes.published.order(:sort_number)
          expect(episodes.count).to eq(3)
          expect(episodes[0].id).to eq(episode1.id)
          expect(episodes[0].raw_number).to eq(nil)
          expect(episodes[1].id).to eq(episode2.id)
          expect(episodes[1].raw_number).to eq(nil)
          expect(episodes[2].id).to eq(episode35.id)
          expect(episodes[2].raw_number).to eq(35.0)
        end
      end

      context "when run on 2019-01-08" do
        before do
          EpisodeGeneratorService.execute!(now: Time.parse("2019-01-08 0:00:00"))
        end

        it "creates 1 episode" do
          episodes = work.episodes.published.order(:sort_number)
          expect(episodes.count).to eq(4)
          expect(episodes[0].id).to eq(episode1.id)
          expect(episodes[0].raw_number).to eq(nil)
          expect(episodes[1].id).to eq(episode2.id)
          expect(episodes[1].raw_number).to eq(nil)
          expect(episodes[2].id).to eq(episode35.id)
          expect(episodes[2].raw_number).to eq(35.0)
          expect(episodes[3].raw_number).to eq(36.0)
        end
      end
    end
  end
end
