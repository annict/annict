# frozen_string_literal: true

describe UserEpisodesQuery, type: :query do
  let!(:user) { create :user }
  let!(:setting) { create :setting, user: user }
  let!(:work_1) { create :work }
  let!(:work_2) { create :work }
  let!(:work_3) { create :work }
  let!(:episode_1) { create :episode, work: work_1 }
  let!(:episode_2) { create :episode, work: work_1 }
  let!(:episode_3) { create :episode, work: work_2 }
  let!(:episode_4) { create :episode, work: work_3 }

  context "when the user does not set status on works" do
    context "when the `watched` option is not specified" do
      it "returns no episodes" do
        episodes = UserEpisodesQuery.new(
          user,
          Episode.all
        ).call

        expect(episodes.pluck(:id)).to match([])
      end
    end

    context "when the `watched` option is `true`" do
      it "returns no episodes" do
        episodes = UserEpisodesQuery.new(
          user,
          Episode.all,
          watched: true
        ).call

        expect(episodes.pluck(:id)).to match([])
      end
    end

    context "when the `watched` option is `false`" do
      it "returns no episodes" do
        episodes = UserEpisodesQuery.new(
          user,
          Episode.all,
          watched: false
        ).call

        expect(episodes.pluck(:id)).to match([])
      end
    end
  end

  context "when the user sets status of work_1 to watching" do
    let!(:status) { create :status, user: user, work: work_1, kind: :watching }

    context "when the user does not track episodes" do
      context "when the `watched` option is not specified" do
        it "returns episodes" do
          episodes = UserEpisodesQuery.new(
            user,
            Episode.all
          ).call

          expect(episodes.pluck(:id)).to contain_exactly(episode_1.id, episode_2.id)
        end
      end

      context "when the `watched` option is `true`" do
        it "returns no episodes" do
          episodes = UserEpisodesQuery.new(
            user,
            Episode.all,
            watched: true
          ).call

          expect(episodes.pluck(:id)).to match([])
        end
      end

      context "when the `watched` option is `false`" do
        it "returns episodes" do
          episodes = UserEpisodesQuery.new(
            user,
            Episode.all,
            watched: false
          ).call

          expect(episodes.pluck(:id)).to contain_exactly(episode_1.id, episode_2.id)
        end
      end
    end

    context "when the user tracks episode_1 which belongs to work_1" do
      let(:episode_record) { build :episode_record, user: user, episode: episode_1 }

      before do
        CreateEpisodeRecordService.new(user, episode_record).save!
      end

      context "when the `watched` option is not specified" do
        it "returns episodes" do
          episodes = UserEpisodesQuery.new(
            user,
            Episode.all
          ).call

          expect(episodes.pluck(:id)).to contain_exactly(episode_1.id, episode_2.id)
        end
      end

      context "when the `watched` option is `true`" do
        it "returns episodes" do
          episodes = UserEpisodesQuery.new(
            user,
            Episode.all,
            watched: true
          ).call

          expect(episodes.pluck(:id)).to contain_exactly(episode_1.id)
        end
      end

      context "when the `watched` option is `false`" do
        it "returns episodes" do
          episodes = UserEpisodesQuery.new(
            user,
            Episode.all,
            watched: false
          ).call

          expect(episodes.pluck(:id)).to contain_exactly(episode_2.id)
        end
      end
    end

    context "when the user tracks episode_3 which belongs to work_2 which is not set status" do
      let(:episode_record) { build :episode_record, user: user, episode: episode_3 }

      before do
        CreateEpisodeRecordService.new(user, episode_record).save!
      end

      context "when the `watched` option is not specified" do
        it "returns episodes" do
          episodes = UserEpisodesQuery.new(
            user,
            Episode.all
          ).call

          expect(episodes.pluck(:id)).to contain_exactly(episode_1.id, episode_2.id)
        end
      end

      context "when the `watched` option is `true`" do
        it "returns no episodes" do
          episodes = UserEpisodesQuery.new(
            user,
            Episode.all,
            watched: true
          ).call

          expect(episodes.pluck(:id)).to match([])
        end
      end

      context "when the `watched` option is `false`" do
        it "returns episodes" do
          episodes = UserEpisodesQuery.new(
            user,
            Episode.all,
            watched: false
          ).call

          expect(episodes.pluck(:id)).to contain_exactly(episode_1.id, episode_2.id)
        end
      end
    end
  end

  context "when the user sets status of work_1 and work_2 to watching" do
    let!(:status_1) { create :status, user: user, work: work_1, kind: :watching }
    let!(:status_2) { create :status, user: user, work: work_2, kind: :watching }

    context "when the user does not track episodes" do
      context "when the `watched` option is not specified" do
        it "returns episodes" do
          episodes = UserEpisodesQuery.new(
            user,
            Episode.all
          ).call

          expect(episodes.pluck(:id)).to contain_exactly(episode_1.id, episode_2.id, episode_3.id)
        end
      end

      context "when the `watched` option is `true`" do
        it "returns no episodes" do
          episodes = UserEpisodesQuery.new(
            user,
            Episode.all,
            watched: true
          ).call

          expect(episodes.pluck(:id)).to match([])
        end
      end

      context "when the `watched` option is `false`" do
        it "returns episodes" do
          episodes = UserEpisodesQuery.new(
            user,
            Episode.all,
            watched: false
          ).call

          expect(episodes.pluck(:id)).to contain_exactly(episode_1.id, episode_2.id, episode_3.id)
        end
      end
    end

    context "when the user tracks episode_1 which belongs to work_1" do
      let(:episode_record) { build :episode_record, user: user, episode: episode_1 }

      before do
        CreateEpisodeRecordService.new(user, episode_record).save!
      end

      context "when the `watched` option is not specified" do
        it "returns episodes" do
          episodes = UserEpisodesQuery.new(
            user,
            Episode.all
          ).call

          expect(episodes.pluck(:id)).to contain_exactly(episode_1.id, episode_2.id, episode_3.id)
        end
      end

      context "when the `watched` option is `true`" do
        it "returns episodes" do
          episodes = UserEpisodesQuery.new(
            user,
            Episode.all,
            watched: true
          ).call

          expect(episodes.pluck(:id)).to contain_exactly(episode_1.id)
        end
      end

      context "when the `watched` option is `false`" do
        it "returns episodes" do
          episodes = UserEpisodesQuery.new(
            user,
            Episode.all,
            watched: false
          ).call

          expect(episodes.pluck(:id)).to contain_exactly(episode_2.id, episode_3.id)
        end
      end
    end
  end

  context "when the user sets status of work_1 to watching and work_2 to dropped" do
    let!(:status_1) { create :status, user: user, work: work_1, kind: :watching }
    let!(:status_2) { create :status, user: user, work: work_2, kind: :stop_watching }

    context "when the user does not track episodes" do
      context "when the `watched` option is not specified" do
        it "returns episodes" do
          episodes = UserEpisodesQuery.new(
            user,
            Episode.all
          ).call

          expect(episodes.pluck(:id)).to contain_exactly(episode_1.id, episode_2.id, episode_3.id)
        end
      end

      context "when the `watched` option is `true`" do
        it "returns no episodes" do
          episodes = UserEpisodesQuery.new(
            user,
            Episode.all,
            watched: true
          ).call

          expect(episodes.pluck(:id)).to match([])
        end
      end

      context "when the `watched` option is `false`" do
        it "returns episodes" do
          episodes = UserEpisodesQuery.new(
            user,
            Episode.all,
            watched: false
          ).call

          expect(episodes.pluck(:id)).to contain_exactly(episode_1.id, episode_2.id, episode_3.id)
        end
      end
    end

    context "when the user tracks episode_1 which belongs to work_1" do
      let(:episode_record) { build :episode_record, user: user, episode: episode_1 }

      before do
        CreateEpisodeRecordService.new(user, episode_record).save!
      end

      context "when the `watched` option is not specified" do
        it "returns episodes" do
          episodes = UserEpisodesQuery.new(
            user,
            Episode.all
          ).call

          expect(episodes.pluck(:id)).to contain_exactly(episode_1.id, episode_2.id, episode_3.id)
        end
      end

      context "when the `watched` option is `true`" do
        it "returns episodes" do
          episodes = UserEpisodesQuery.new(
            user,
            Episode.all,
            watched: true
          ).call

          expect(episodes.pluck(:id)).to contain_exactly(episode_1.id)
        end
      end

      context "when the `watched` option is `false`" do
        it "returns episodes" do
          episodes = UserEpisodesQuery.new(
            user,
            Episode.all,
            watched: false
          ).call

          expect(episodes.pluck(:id)).to contain_exactly(episode_2.id, episode_3.id)
        end
      end
    end

    context "when the user tracks episode_3 which belongs to work_2" do
      let(:episode_record) { build :episode_record, user: user, episode: episode_3 }

      before do
        CreateEpisodeRecordService.new(user, episode_record).save!
      end

      context "when the `watched` option is not specified" do
        it "returns episodes" do
          episodes = UserEpisodesQuery.new(
            user,
            Episode.all
          ).call

          expect(episodes.pluck(:id)).to contain_exactly(episode_1.id, episode_2.id, episode_3.id)
        end
      end

      context "when the `watched` option is `true`" do
        it "returns no episodes" do
          episodes = UserEpisodesQuery.new(
            user,
            Episode.all,
            watched: true
          ).call

          expect(episodes.pluck(:id)).to contain_exactly(episode_3.id)
        end
      end

      context "when the `watched` option is `false`" do
        it "returns episodes" do
          episodes = UserEpisodesQuery.new(
            user,
            Episode.all,
            watched: false
          ).call

          expect(episodes.pluck(:id)).to contain_exactly(episode_1.id, episode_2.id)
        end
      end
    end
  end
end
