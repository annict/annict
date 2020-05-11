# frozen_string_literal: true

describe CreateWorkRecordService, type: :service do
  describe do
    let(:user) { create :registered_user }
    let(:work) { create :work }

    it "creates work record" do
      expect(WorkRecord.count).to eq 0
      expect(Record.count).to eq 0
      expect(Activity.count).to eq 0
      expect(work.work_records_with_body_count).to eq 0
      expect(user.share_record_to_twitter?).to eq false

      attrs = {
        body: "すごく面白かった。",
        rating_overall_state: "great",
        rating_animation_state: "great",
        rating_character_state: "great",
        rating_music_state: "great",
        rating_story_state: "great"
      }
      CreateWorkRecordService.new(user: user, work: work).call(attrs)

      expect(WorkRecord.count).to eq 1
      expect(Record.count).to eq 1
      expect(Activity.count).to eq 1
      expect(work.work_records_with_body_count).to eq 1
      expect(user.share_record_to_twitter?).to eq false

      work_record = user.work_records.first
      record = user.records.first
      activity = user.activities.first

      expect(work_record.body).to eq attrs[:body]
      expect(work_record.locale).to eq "ja"
      expect(work_record.rating_overall_state).to eq attrs[:rating_overall_state]
      expect(work_record.rating_animation_state).to eq attrs[:rating_animation_state]
      expect(work_record.rating_character_state).to eq attrs[:rating_character_state]
      expect(work_record.rating_music_state).to eq attrs[:rating_music_state]
      expect(work_record.rating_story_state).to eq attrs[:rating_story_state]
      expect(work_record.activity_id).to eq activity.id
      expect(work_record.record_id).to eq record.id
      expect(work_record.work_id).to eq work.id

      expect(record.work_id).to eq work.id

      expect(activity.action).to eq "create_work_record"
      expect(activity.single).to eq true
      expect(activity.trackable_type).to eq "WorkRecord"
    end
  end
end
