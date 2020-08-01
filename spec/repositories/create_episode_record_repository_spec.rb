# frozen_string_literal: true

describe CreateEpisodeRecordRepository, type: :repository do
  include V4::GraphqlRunnable

  describe do
    let(:user) { create :registered_user }
    let(:episode) { create :episode }
    let(:work) { episode.work }

    it "creates episode record" do
      expect(Record.count).to eq 0
      expect(EpisodeRecord.count).to eq 0
      expect(ActivityGroup.count).to eq 0
      expect(Activity.count).to eq 0
      expect(user.share_record_to_twitter?).to eq false

      params = {
        body: "すごく面白かった。",
        rating_state: "great"
      }
      CreateEpisodeRecordRepository.new(
        graphql_client: graphql_client(viewer: user)
      ).create(episode: episode, params: params)

      expect(Record.count).to eq 1
      expect(EpisodeRecord.count).to eq 1
      expect(ActivityGroup.count).to eq 1
      expect(Activity.count).to eq 1
      expect(user.share_record_to_twitter?).to eq false

      record = user.records.first
      episode_record = user.episode_records.first
      activity_group = user.activity_groups.first
      activity = user.activities.first

      expect(record.anime_id).to eq work.id

      expect(episode_record.body).to eq params[:body]
      expect(episode_record.locale).to eq "ja"
      expect(episode_record.rating_state).to eq params[:rating_state]
      expect(episode_record.episode_id).to eq episode.id
      expect(episode_record.record_id).to eq record.id
      expect(episode_record.anime_id).to eq work.id

      expect(activity_group.itemable_type).to eq "EpisodeRecord"
      expect(activity_group.single).to eq true

      expect(activity.activity_group_id).to eq activity_group.id
      expect(activity.itemable).to eq episode_record
    end
  end

  context "when episode record with body has been created and create new episode record with body" do
    let(:user) { create :registered_user }
    let(:episode) { create :episode, episode_record_bodies_count: 1 }
    let(:work) { episode.work }
    let!(:episode_record) { create(:episode_record, user: user, episode: episode, body: "さいこー") }
    let!(:activity_group) { create(:activity_group, user: user, itemable_type: "EpisodeRecord", single: true) }
    let!(:activity) { create(:activity, user: user, activity_group: activity_group, itemable: episode_record) }

    it "creates episode record" do
      expect(Record.count).to eq 1
      expect(EpisodeRecord.count).to eq 1
      expect(ActivityGroup.count).to eq 1
      expect(Activity.count).to eq 1
      expect(user.share_record_to_twitter?).to eq false

      params = {
        body: "すごく面白かった。",
        rating_state: "great"
      }
      CreateEpisodeRecordRepository.new(
        graphql_client: graphql_client(viewer: user)
      ).create(episode: episode, params: params)

      expect(Record.count).to eq 2
      expect(EpisodeRecord.count).to eq 2
      expect(ActivityGroup.count).to eq 2
      expect(Activity.count).to eq 2
      expect(user.share_record_to_twitter?).to eq false

      record = user.records.last
      episode_record = user.episode_records.last
      activity_group = user.activity_groups.last
      activity = user.activities.last

      expect(record.anime_id).to eq work.id

      expect(episode_record.body).to eq params[:body]
      expect(episode_record.locale).to eq "ja"
      expect(episode_record.rating_state).to eq params[:rating_state]
      expect(episode_record.episode_id).to eq episode.id
      expect(episode_record.record_id).to eq record.id
      expect(episode_record.anime_id).to eq work.id

      expect(activity_group.itemable_type).to eq "EpisodeRecord"
      expect(activity_group.single).to eq true

      expect(activity.activity_group_id).to eq activity_group.id
      expect(activity.itemable).to eq episode_record
    end
  end

  context "when episode record without body has been created and create new episode record without body" do
    let(:user) { create :registered_user }
    let(:episode) { create :episode }
    let(:work) { episode.work }
    let!(:episode_record) { create(:episode_record, user: user, episode: episode, body: "") }
    let!(:activity_group) { create(:activity_group, user: user, itemable_type: "EpisodeRecord", single: false) }
    let!(:activity) { create(:activity, user: user, activity_group: activity_group, itemable: episode_record) }

    it "creates episode record" do
      expect(Record.count).to eq 1
      expect(EpisodeRecord.count).to eq 1
      expect(ActivityGroup.count).to eq 1
      expect(Activity.count).to eq 1
      expect(user.share_record_to_twitter?).to eq false

      params = {
        body: "",
        rating_state: "great"
      }
      CreateEpisodeRecordRepository.new(
        graphql_client: graphql_client(viewer: user)
      ).create(episode: episode, params: params)

      expect(Record.count).to eq 2
      expect(EpisodeRecord.count).to eq 2
      expect(ActivityGroup.count).to eq 1
      expect(Activity.count).to eq 2
      expect(user.share_record_to_twitter?).to eq false

      record = user.records.last
      episode_record = user.episode_records.last
      activity_group = user.activity_groups.first
      activity = user.activities.last

      expect(episode_record.body).to eq ""
      expect(episode_record.locale).to eq "other"
      expect(episode_record.rating_state).to eq params[:rating_state]
      expect(episode_record.episode_id).to eq episode.id
      expect(episode_record.record_id).to eq record.id
      expect(episode_record.anime_id).to eq work.id

      expect(activity_group.itemable_type).to eq "EpisodeRecord"
      expect(activity_group.single).to eq false

      expect(activity.activity_group_id).to eq activity_group.id
      expect(activity.itemable).to eq episode_record
    end
  end
end
