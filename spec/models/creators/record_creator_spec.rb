# frozen_string_literal: true

describe Creators::RecordCreator, type: :model do
  let!(:user) { create :registered_user }
  let!(:work) { create :work }

  context "作品の記録をするとき" do
    it "記録ができること" do
      # Creatorを呼んでいないので、各レコードは0件のはず
      expect(Record.count).to eq 0
      expect(ActivityGroup.count).to eq 0
      expect(Activity.count).to eq 0
      expect(user.share_record_to_twitter?).to eq false

      # Creatorを呼ぶ
      Creators::RecordCreator.new(
        user: user,
        form: Forms::RecordForm.new(
          work_id: work.id,
          body: "すごく面白かった。",
          rating: "great",
          share_to_twitter: false
        )
      ).call

      # Creatorを呼んだので、各レコードが1件ずつ作成されるはず
      expect(Record.count).to eq 1
      expect(ActivityGroup.count).to eq 1
      expect(Activity.count).to eq 1
      expect(user.share_record_to_twitter?).to eq false

      record = user.records.first
      activity_group = user.activity_groups.first
      activity = user.activities.first

      expect(record.work_id).to eq work.id

      expect(record.body).to eq "すごく面白かった。"
      expect(record.locale).to eq "ja"
      expect(record.rating).to eq "great"
      expect(record.work_id).to eq work.id

      expect(activity_group.itemable_type).to eq "Record"
      expect(activity_group.single).to eq true

      expect(activity.itemable).to eq record
      expect(activity.activity_group_id).to eq activity_group.id
    end
  end

  context "エピソードの記録をするとき" do
    let!(:episode) { create :episode, work: work, sort_number: 10 }
    let!(:next_episode) { create :episode, work: work, sort_number: 20 }

    it "記録ができること" do
      # Creatorを呼んでいないので、各レコードは0件のはず
      expect(Record.count).to eq 0
      expect(ActivityGroup.count).to eq 0
      expect(Activity.count).to eq 0
      expect(LibraryEntry.count).to eq 0
      expect(user.share_record_to_twitter?).to eq false

      # Creatorを呼ぶ
      Creators::RecordCreator.new(
        user: user,
        form: Forms::RecordForm.new(
          body: "にぱー",
          episode_id: episode.id,
          rating: "good",
          share_to_twitter: false
        )
      ).call

      # Creatorを呼んだので、各レコードが1件ずつ作成されるはず
      expect(Record.count).to eq 1
      expect(ActivityGroup.count).to eq 1
      expect(Activity.count).to eq 1
      expect(LibraryEntry.count).to eq 1
      expect(user.share_record_to_twitter?).to eq false

      record = user.records.first
      activity_group = user.activity_groups.first
      activity = user.activities.first
      library_entry = user.library_entries.first

      expect(record.work_id).to eq work.id
      expect(record.episode_id).to eq episode.id
      expect(record.body).to eq "にぱー"
      expect(record.locale).to eq "ja"
      expect(record.rating).to eq "good"

      expect(activity_group.itemable_type).to eq "Record"
      expect(activity_group.single).to eq true

      expect(activity.activity_group_id).to eq activity_group.id
      expect(activity.itemable).to eq record

      expect(library_entry.work).to eq work
      expect(library_entry.watched_episode_ids).to eq [episode.id]
      expect(library_entry.next_episode).to eq next_episode
    end
  end

  describe "アクティビティの作成" do
    context "直前の記録に感想が書かれていて、その後に新たに感想付きの記録をしたとき" do
      let(:episode) { create :episode, episode_record_bodies_count: 1 }
      let(:work) { episode.work }
      # 感想付きの記録が直前にある
      let(:record) { create(:record, user: user, episode: episode, body: "はうー") }
      let!(:activity_group) { create(:activity_group, user: user, itemable_type: "Record", single: true) }
      let!(:activity) { create(:activity, user: user, activity_group: activity_group, itemable: record) }

      it "ActivityGroup が新たに作成されること" do
        expect(Record.count).to eq 1
        expect(ActivityGroup.count).to eq 1
        expect(Activity.count).to eq 1
        expect(user.share_record_to_twitter?).to eq false

        # Creatorを呼ぶ
        Creators::RecordCreator.new(
          user: user,
          form: Forms::RecordForm.new(
            body: "にぱー", # 感想付きの記録を新たにする
            episode_id: episode.id,
            rating: "good",
            share_to_twitter: false
          )
        ).call

        expect(Record.count).to eq 2
        expect(ActivityGroup.count).to eq 2 # ActivityGroup が新たに作成されるはず
        expect(Activity.count).to eq 2

        record = user.records.last
        activity_group = user.activity_groups.last
        activity = user.activities.last

        expect(activity_group.itemable_type).to eq "Record"
        expect(activity_group.single).to eq true

        expect(activity.activity_group_id).to eq activity_group.id
        expect(activity.itemable).to eq record
      end
    end

    context "直前の記録に感想が書かれていない & その後に新たに感想無しの記録をしたとき" do
      let(:user) { create :registered_user }
      let(:episode) { create :episode }
      let(:work) { episode.work }
      # 感想無しの記録が直前にある
      let(:record) { create(:record, user: user, episode: episode, body: "") }
      let!(:activity_group) { create(:activity_group, user: user, itemable_type: "Record", single: false) }
      let!(:activity) { create(:activity, user: user, activity_group: activity_group, itemable: record) }

      it "ActivityGroup が新たに作成されないこと" do
        expect(Record.count).to eq 1
        expect(ActivityGroup.count).to eq 1
        expect(Activity.count).to eq 1
        expect(user.share_record_to_twitter?).to eq false

        # Creatorを呼ぶ
        Creators::RecordCreator.new(
          user: user,
          form: Forms::RecordForm.new(
            body: "", # 感想無しの記録を新たにする
            episode_id: episode.id,
            rating: "good",
            share_to_twitter: false
          )
        ).call

        expect(Record.count).to eq 2
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
