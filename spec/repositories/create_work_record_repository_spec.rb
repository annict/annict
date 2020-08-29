# frozen_string_literal: true

describe CreateAnimeRecordRepository, type: :repository do
  include V4::GraphqlRunnable

  describe do
    let(:user) { create :registered_user }
    let(:work) { create :work }

    it "creates work record" do
      expect(Record.count).to eq 0
      expect(WorkRecord.count).to eq 0
      expect(ActivityGroup.count).to eq 0
      expect(Activity.count).to eq 0
      expect(user.share_record_to_twitter?).to eq false

      params = {
        body: "すごく面白かった。",
        rating_overall_state: "great",
        rating_animation_state: "great",
        rating_character_state: "great",
        rating_music_state: "great",
        rating_story_state: "great"
      }
      CreateAnimeRecordRepository.new(graphql_client: graphql_client(viewer: user)).execute(anime: work, params: params)

      expect(Record.count).to eq 1
      expect(WorkRecord.count).to eq 1
      expect(ActivityGroup.count).to eq 1
      expect(Activity.count).to eq 1
      expect(user.share_record_to_twitter?).to eq false

      record = user.records.first
      work_record = user.work_records.first
      activity_group = user.activity_groups.first
      activity = user.activities.first

      expect(record.work_id).to eq work.id

      expect(work_record.body).to eq params[:body]
      expect(work_record.locale).to eq "ja"
      expect(work_record.rating_overall_state).to eq params[:rating_overall_state]
      expect(work_record.rating_animation_state).to eq params[:rating_animation_state]
      expect(work_record.rating_character_state).to eq params[:rating_character_state]
      expect(work_record.rating_music_state).to eq params[:rating_music_state]
      expect(work_record.rating_story_state).to eq params[:rating_story_state]
      expect(work_record.record_id).to eq record.id
      expect(work_record.work_id).to eq work.id

      expect(activity_group.itemable_type).to eq "WorkRecord"
      expect(activity_group.single).to eq true

      expect(activity.itemable).to eq work_record
      expect(activity.activity_group_id).to eq activity_group.id
    end
  end

  context "when episode record with body has been created and create new episode record with body" do
    let(:user) { create :registered_user }
    let(:work) { create :work, work_records_with_body_count: 1 }
    let!(:work_record) { create(:work_record, user: user, work: work, body: "さいこー") }
    let!(:activity_group) { create(:activity_group, user: user, itemable_type: "WorkRecord", single: true) }
    let!(:activity) { create(:activity, user: user, itemable: work_record, activity_group: activity_group) }

    it "creates work record" do
      expect(Record.count).to eq 1
      expect(WorkRecord.count).to eq 1
      expect(ActivityGroup.count).to eq 1
      expect(Activity.count).to eq 1
      expect(user.share_record_to_twitter?).to eq false

      params = {
        body: "すごく面白かった。",
        rating_overall_state: "great",
        rating_animation_state: "great",
        rating_character_state: "great",
        rating_music_state: "great",
        rating_story_state: "great"
      }
      CreateAnimeRecordRepository.new(graphql_client: graphql_client(viewer: user)).execute(anime: work, params: params)

      expect(Record.count).to eq 2
      expect(WorkRecord.count).to eq 2
      expect(ActivityGroup.count).to eq 2
      expect(Activity.count).to eq 2
      expect(user.share_record_to_twitter?).to eq false

      record = user.records.last
      work_record = user.work_records.last
      activity_group = user.activity_groups.last
      activity = user.activities.last

      expect(record.work_id).to eq work.id

      expect(work_record.body).to eq params[:body]
      expect(work_record.locale).to eq "ja"
      expect(work_record.rating_overall_state).to eq params[:rating_overall_state]
      expect(work_record.rating_animation_state).to eq params[:rating_animation_state]
      expect(work_record.rating_character_state).to eq params[:rating_character_state]
      expect(work_record.rating_music_state).to eq params[:rating_music_state]
      expect(work_record.rating_story_state).to eq params[:rating_story_state]
      expect(work_record.record_id).to eq record.id
      expect(work_record.work_id).to eq work.id

      expect(activity_group.itemable_type).to eq "WorkRecord"
      expect(activity_group.single).to eq true

      expect(activity.itemable).to eq work_record
      expect(activity.activity_group_id).to eq activity_group.id
    end
  end

  context "when work record without body has been created and create new work record without body" do
    let(:user) { create :registered_user }
    let(:work) { create :work }
    let!(:work_record) { create(:work_record, user: user, work: work, body: "") }
    let!(:activity_group) { create(:activity_group, user: user, itemable_type: "WorkRecord", single: false) }
    let!(:activity) { create(:activity, user: user, itemable: work_record, activity_group: activity_group) }

    it "creates work record" do
      expect(Record.count).to eq 1
      expect(WorkRecord.count).to eq 1
      expect(ActivityGroup.count).to eq 1
      expect(Activity.count).to eq 1
      expect(user.share_record_to_twitter?).to eq false

      params = {
        body: "",
        rating_overall_state: "great",
        rating_animation_state: "great",
        rating_character_state: "great",
        rating_music_state: "great",
        rating_story_state: "great"
      }
      CreateAnimeRecordRepository.new(graphql_client: graphql_client(viewer: user)).execute(anime: work, params: params)

      expect(Record.count).to eq 2
      expect(WorkRecord.count).to eq 2
      expect(ActivityGroup.count).to eq 1
      expect(Activity.count).to eq 2
      expect(user.share_record_to_twitter?).to eq false

      record = user.records.last
      work_record = user.work_records.last
      activity_group = user.activity_groups.first
      activity = user.activities.last

      expect(record.work_id).to eq work.id

      expect(work_record.body).to eq params[:body]
      expect(work_record.locale).to eq "other"
      expect(work_record.rating_overall_state).to eq params[:rating_overall_state]
      expect(work_record.rating_animation_state).to eq params[:rating_animation_state]
      expect(work_record.rating_character_state).to eq params[:rating_character_state]
      expect(work_record.rating_music_state).to eq params[:rating_music_state]
      expect(work_record.rating_story_state).to eq params[:rating_story_state]
      expect(work_record.record_id).to eq record.id
      expect(work_record.work_id).to eq work.id

      expect(activity_group.itemable_type).to eq "WorkRecord"
      expect(activity_group.single).to eq false

      expect(activity.itemable).to eq work_record
      expect(activity.activity_group_id).to eq activity_group.id
    end
  end
end
