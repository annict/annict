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
      expect(activity.recipient).to eq episode
      expect(activity.trackable).to eq episode_record
      expect(activity.single).to eq true
      expect(activity.repetitiveness).to eq false
    end
  end

  context "when episode record with body has been created and create new episode record with body" do
    let(:user) { create :registered_user }
    let(:episode) { create :episode, episode_record_bodies_count: 1 }
    let(:work) { episode.work }
    let!(:episode_record) { create(:episode_record, user: user, episode: episode, body: "さいこー") }
    let!(:activity) { create(:activity, user: user, recipient: episode, trackable: episode_record, action: :create_episode_record, single: true) }

    it "creates episode record" do
      expect(EpisodeRecord.count).to eq 1
      expect(Record.count).to eq 1
      expect(Activity.count).to eq 1
      expect(episode.episode_record_bodies_count).to eq 1
      expect(user.share_record_to_twitter?).to eq false

      attrs = {
        body: "すごく面白かった。",
        rating_state: "great"
      }
      CreateEpisodeRecordService.new(user: user, episode: episode).call(attrs)

      expect(EpisodeRecord.count).to eq 2
      expect(Record.count).to eq 2
      expect(Activity.count).to eq 2
      expect(episode.episode_record_bodies_count).to eq 2
      expect(user.share_record_to_twitter?).to eq false

      episode_record = user.episode_records.last
      record = user.records.last
      activity = user.activities.last

      expect(episode_record.body).to eq attrs[:body]
      expect(episode_record.locale).to eq "ja"
      expect(episode_record.rating_state).to eq attrs[:rating_state]
      expect(episode_record.activity_id).to eq activity.id
      expect(episode_record.episode_id).to eq episode.id
      expect(episode_record.record_id).to eq record.id
      expect(episode_record.work_id).to eq work.id

      expect(record.work_id).to eq work.id

      expect(activity.action).to eq "create_episode_record"
      expect(activity.recipient).to eq episode
      expect(activity.trackable).to eq episode_record
      expect(activity.single).to eq true
      expect(activity.repetitiveness).to eq false
    end
  end

  context "when episode record without body has been created and create new episode record without body" do
    let(:user) { create :registered_user }
    let(:episode) { create :episode }
    let(:work) { episode.work }
    let!(:episode_record) { create(:episode_record, user: user, episode: episode, body: "") }
    let!(:activity) { create(:activity, user: user, recipient: episode, trackable: episode_record, action: :create_episode_record, single: false) }

    it "creates episode record" do
      expect(EpisodeRecord.count).to eq 1
      expect(Record.count).to eq 1
      expect(Activity.count).to eq 1
      expect(episode.episode_record_bodies_count).to eq 0
      expect(user.share_record_to_twitter?).to eq false

      attrs = {
        body: "",
        rating_state: "great"
      }
      CreateEpisodeRecordService.new(user: user, episode: episode).call(attrs)

      expect(EpisodeRecord.count).to eq 2
      expect(Record.count).to eq 2
      expect(Activity.count).to eq 2
      expect(episode.episode_record_bodies_count).to eq 0
      expect(user.share_record_to_twitter?).to eq false

      episode_record = user.episode_records.last
      record = user.records.last
      activity_1 = user.activities.first
      activity_2 = user.activities.last

      expect(episode_record.body).to eq ""
      expect(episode_record.locale).to eq "other"
      expect(episode_record.rating_state).to eq attrs[:rating_state]
      expect(episode_record.activity_id).to eq activity_1.id
      expect(episode_record.episode_id).to eq episode.id
      expect(episode_record.record_id).to eq record.id
      expect(episode_record.work_id).to eq work.id

      expect(record.work_id).to eq work.id

      expect(activity_2.action).to eq "create_episode_record"
      expect(activity_2.recipient).to eq episode
      expect(activity_2.trackable).to eq episode_record
      expect(activity_2.single).to eq false
      expect(activity_2.repetitiveness).to eq true
    end
  end
end
