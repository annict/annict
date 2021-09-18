# frozen_string_literal: true

describe Creators::WorkRecordCreator, type: :model do
  let!(:current_time) { Time.zone.parse("2021-09-01 10:00:00") }

  before do
    travel_to current_time
  end

  context "watched_at が指定されていないとき" do
    let!(:user) { create :registered_user }
    let!(:work) { create :work }

    it "作品への記録ができること" do
      # Creatorを呼んでいないので、各レコードは0件のはず
      expect(Record.count).to eq 0
      expect(WorkRecord.count).to eq 0
      expect(ActivityGroup.count).to eq 0
      expect(Activity.count).to eq 0
      expect(user.share_record_to_twitter?).to eq false

      # Creatorを呼ぶ
      Creators::WorkRecordCreator.new(
        user: user,
        form: WorkRecordForm.new(
          user: user,
          work: work,
          body: "すごく面白かった。",
          animation_rating: "great",
          character_rating: "great",
          music_rating: "great",
          story_rating: "great",
          rating: "great",
          share_to_twitter: false
        )
      ).call

      # Creatorを呼んだので、各レコードが1件ずつ作成されるはず
      expect(Record.count).to eq 1
      expect(WorkRecord.count).to eq 1
      expect(ActivityGroup.count).to eq 1
      expect(Activity.count).to eq 1
      expect(user.share_record_to_twitter?).to eq false

      record = user.records.first
      work_record = record.work_record
      activity_group = user.activity_groups.first
      activity = user.activities.first

      expect(record.work_id).to eq work.id
      expect(record.body).to eq "すごく面白かった。"
      expect(record.locale).to eq "ja"
      expect(record.rating).to eq "great"
      expect(record.animation_rating).to eq "great"
      expect(record.character_rating).to eq "great"
      expect(record.music_rating).to eq "great"
      expect(record.story_rating).to eq "great"

      expect(work_record).to_not be_nil

      expect(activity_group.itemable_type).to eq "Record"
      expect(activity_group.single).to eq true

      expect(activity.itemable).to eq record
      expect(activity.activity_group_id).to eq activity_group.id
    end
  end

  context "watched_at が指定されているとき" do
    let!(:watched_time) { Time.zone.parse("2021-01-01 12:00:00") }
    let!(:user) { create :registered_user }
    let!(:work) { create :work }

    it "作品への記録が作成できること" do
      Creators::WorkRecordCreator.new(
        user: user,
        form: WorkRecordForm.new(
          user: user,
          work: work,
          watched_at: watched_time
        )
      ).call

      record = user.records.first

      # watched_at が指定した日時になっているはず
      expect(record.watched_at).to eq watched_time

      # ActivityGroup, Activity は作成されないはず
      expect(ActivityGroup.count).to eq 0
      expect(Activity.count).to eq 0
    end
  end

  describe "アクティビティの作成" do
    context "直前の記録に感想が書かれていて、その後に新たに感想付きの記録をしたとき" do
      let!(:user) { create :registered_user }
      let!(:work) { create :work, work_records_with_body_count: 1 }
      # 感想付きの記録が直前にある
      let!(:work_record) { create(:work_record) }
      let!(:record) { create(:record, :on_work, user: user, work: work, recordable: work_record, body: "さいこー") }
      let!(:activity_group) { create(:activity_group, user: user, itemable_type: "Record", single: true) }
      let!(:activity) { create(:activity, user: user, itemable: record, activity_group: activity_group) }

      it "ActivityGroup が新たに作成されること" do
        expect(ActivityGroup.count).to eq 1
        expect(Activity.count).to eq 1

        # Creatorを呼ぶ
        Creators::WorkRecordCreator.new(
          user: user,
          form: WorkRecordForm.new(
            user: user,
            work: work,
            body: "すごく面白かった。", # 感想付きの記録を新たにする
            animation_rating: "great",
            character_rating: "great",
            music_rating: "great",
            story_rating: "great",
            rating: "great",
            share_to_twitter: false
          )
        ).call

        expect(ActivityGroup.count).to eq 2 # ActivityGroup が新たに作成されるはず
        expect(Activity.count).to eq 2

        record = user.records.last
        activity_group = user.activity_groups.last
        activity = user.activities.last

        expect(activity_group.itemable_type).to eq "Record"
        expect(activity_group.single).to eq true

        expect(activity.itemable).to eq record
        expect(activity.activity_group_id).to eq activity_group.id
      end
    end

    context "直前の記録に感想が書かれていない & その後に新たに感想無しの記録をしたとき" do
      let!(:user) { create :registered_user }
      let!(:work) { create :work }
      # 感想無しの記録が直前にある
      let!(:work_record) { create(:work_record) }
      let!(:record) { create(:record, :on_work, user: user, work: work, recordable: work_record, body: "") }
      let!(:activity_group) { create(:activity_group, user: user, itemable_type: "Record", single: false) }
      let!(:activity) { create(:activity, user: user, itemable: record, activity_group: activity_group) }

      it "ActivityGroup が新たに作成されないこと" do
        expect(ActivityGroup.count).to eq 1
        expect(Activity.count).to eq 1

        # Creatorを呼ぶ
        Creators::WorkRecordCreator.new(
          user: user,
          form: WorkRecordForm.new(
            user: user,
            work: work,
            body: "", # 感想無しの記録を新たにする
            animation_rating: "great",
            character_rating: "great",
            music_rating: "great",
            story_rating: "great",
            rating: "great",
            share_to_twitter: false
          )
        ).call

        expect(ActivityGroup.count).to eq 1 # ActivityGroup は新たに作成されないはず
        expect(Activity.count).to eq 2

        record = user.records.last
        activity_group = user.activity_groups.first
        activity = user.activities.last

        expect(activity_group.itemable_type).to eq "Record"
        expect(activity_group.single).to eq false

        expect(activity.itemable).to eq record
        # もともとあった ActivityGroup に紐付くはず
        expect(activity.activity_group_id).to eq activity_group.id
      end
    end
  end
end
