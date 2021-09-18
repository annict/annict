# frozen_string_literal: true

describe Creators::EpisodeRecordCreator, type: :model do
  let!(:user) { create :registered_user }
  let!(:work) { create :work }
  let!(:episode) { create :episode, work: work, sort_number: 10 }
  let!(:next_episode) { create :episode, work: work, sort_number: 20 }

  it "エピソードへの記録が作成できること" do
    # Creatorを呼んでいないので、各レコードは0件のはず
    expect(Record.count).to eq 0
    expect(EpisodeRecord.count).to eq 0
    expect(ActivityGroup.count).to eq 0
    expect(Activity.count).to eq 0
    expect(LibraryEntry.count).to eq 0
    expect(user.share_record_to_twitter?).to eq false

    # Creatorを呼ぶ
    Creators::EpisodeRecordCreator.new(
      user: user,
      form: EpisodeRecordForm.new(
        user: user,
        body: "にぱー",
        episode: episode,
        rating: "good",
        share_to_twitter: false
      )
    ).call

    # Creatorを呼んだので、各レコードが1件ずつ作成されるはず
    expect(Record.count).to eq 1
    expect(EpisodeRecord.count).to eq 1
    expect(ActivityGroup.count).to eq 1
    expect(Activity.count).to eq 1
    expect(LibraryEntry.count).to eq 1
    expect(user.share_record_to_twitter?).to eq false

    record = user.records.first
    episode_record = record.episode_record
    activity_group = user.activity_groups.first
    activity = user.activities.first
    library_entry = user.library_entries.first

    expect(record.body).to eq "にぱー"
    expect(record.locale).to eq "ja"
    expect(record.rating).to eq "good"
    expect(record.episode_id).to eq episode.id
    expect(record.work_id).to eq work.id

    expect(episode_record).not_to be_nil

    expect(activity_group.itemable_type).to eq "Record"
    expect(activity_group.single).to eq true

    expect(activity.activity_group_id).to eq activity_group.id
    expect(activity.itemable).to eq record

    expect(library_entry.work).to eq work
    expect(library_entry.watched_episode_ids).to eq [episode.id]
    expect(library_entry.next_episode).to eq next_episode
  end

  describe "アクティビティの作成" do
    context "直前の記録に感想が書かれていて、その後に新たに感想付きの記録をしたとき" do
      let(:episode) { create :episode, episode_record_bodies_count: 1 }
      let(:work) { episode.work }
      # 感想付きの記録が直前にある
      let!(:record) { create(:record, :on_episode, user: user, work: work, episode: episode, body: "はうー") }
      let!(:activity_group) { create(:activity_group, user: user, itemable_type: "Record", single: true) }
      let!(:activity) { create(:activity, user: user, activity_group: activity_group, itemable: record) }

      it "ActivityGroup が新たに作成されること" do
        expect(Record.count).to eq 1
        expect(EpisodeRecord.count).to eq 1
        expect(ActivityGroup.count).to eq 1
        expect(Activity.count).to eq 1
        expect(user.share_record_to_twitter?).to eq false

        # Creatorを呼ぶ
        Creators::EpisodeRecordCreator.new(
          user: user,
          form: EpisodeRecordForm.new(
            user: user,
            body: "にぱー", # 感想付きの記録を新たにする
            episode: episode,
            rating: "good",
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

        expect(activity.activity_group_id).to eq activity_group.id
        expect(activity.itemable).to eq record
      end
    end

    context "直前の記録に感想が書かれていない & その後に新たに感想無しの記録をしたとき" do
      let(:user) { create :registered_user }
      let(:episode) { create :episode }
      let(:work) { episode.work }
      # 感想無しの記録が直前にある
      let!(:record) { create(:record, :on_episode, user: user, work: work, episode: episode, body: "") }
      let!(:activity_group) { create(:activity_group, user: user, itemable_type: "Record", single: false) }
      let!(:activity) { create(:activity, user: user, activity_group: activity_group, itemable: record) }

      it "ActivityGroup が新たに作成されないこと" do
        expect(Record.count).to eq 1
        expect(EpisodeRecord.count).to eq 1
        expect(ActivityGroup.count).to eq 1
        expect(Activity.count).to eq 1
        expect(user.share_record_to_twitter?).to eq false

        # Creatorを呼ぶ
        Creators::EpisodeRecordCreator.new(
          user: user,
          form: EpisodeRecordForm.new(
            user: user,
            body: "", # 感想無しの記録を新たにする
            episode: episode,
            rating: "good",
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
