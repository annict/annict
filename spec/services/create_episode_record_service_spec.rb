# frozen_string_literal: true

describe CreateEpisodeRecordService, type: :service do
  describe do
    let(:user) { create :registered_user }
    let(:episode) { create :episode }
    let(:work) { episode.work }

    it "creates episode record" do
      expect(EpisodeRecord.count).to eq 0
      expect(Record.count).to eq 0
      expect(Activity.count).to eq 0
      expect(episode.episode_record_bodies_count).to eq 0
      expect(user.share_record_to_twitter?).to eq false

      attrs = {
        body: "すごく面白かった。",
        rating_state: "great"
      }
      CreateEpisodeRecordService.new(user: user, episode: episode).call(attrs)

      expect(EpisodeRecord.count).to eq 1
      expect(Record.count).to eq 1
      expect(Activity.count).to eq 1
      expect(episode.episode_record_bodies_count).to eq 1
      expect(user.share_record_to_twitter?).to eq false

      episode_record = user.episode_records.first
      record = user.records.first
      activity = user.activities.first

      expect(episode_record.body).to eq attrs[:body]
      expect(episode_record.locale).to eq "ja"
      expect(episode_record.rating_state).to eq attrs[:rating_state]
      expect(episode_record.activity_id).to eq activity.id
      expect(episode_record.episode_id).to eq episode.id
      expect(episode_record.record_id).to eq record.id
      expect(episode_record.work_id).to eq work.id

      expect(record.work_id).to eq work.id

      expect(activity.action).to eq "create_episode_record"
      expect(activity.solo).to eq true
      expect(activity.trackable_type).to eq "EpisodeRecord"
    end
  end
end
