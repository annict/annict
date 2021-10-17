# frozen_string_literal: true

describe Creators::EpisodeRecordCreator, type: :model do
  let!(:current_time) { Time.zone.parse("2021-09-01 10:00:00") }
  let!(:user) { create :registered_user, :with_supporter }
  let!(:work) { create :work }
  let!(:episode) { create :episode, work: work, sort_number: 10 }
  let!(:next_episode) { create :episode, work: work, sort_number: 20 }

  before do
    travel_to current_time
  end

  it "エピソードへの記録が作成できること" do
    # Creatorを呼んでいないので、各レコードは0件のはず
    expect(Record.count).to eq 0
    expect(EpisodeRecord.count).to eq 0
    expect(ActivityGroup.count).to eq 0
    expect(Activity.count).to eq 0
    expect(LibraryEntry.count).to eq 0
    expect(user.share_record_to_twitter?).to eq false

    # Creatorを呼ぶ
    form = Forms::EpisodeRecordForm.new(user: user, episode: episode)
    form.attributes = {
      comment: "にぱー",
      rating: "good",
      share_to_twitter: false
    }
    expect(form.valid?).to eq true

    Creators::EpisodeRecordCreator.new(user: user, form: form).call

    # Creatorを呼んだので、各レコードが1件ずつ作成されるはず
    expect(Record.count).to eq 1
    expect(EpisodeRecord.count).to eq 1
    expect(ActivityGroup.count).to eq 1
    expect(Activity.count).to eq 1
    expect(LibraryEntry.count).to eq 1
    expect(user.share_record_to_twitter?).to eq false

    record = user.records.first
    episode_record = user.episode_records.first
    activity_group = user.activity_groups.first
    activity = user.activities.first
    library_entry = user.library_entries.first

    expect(record.work_id).to eq work.id
    expect(record.watched_at).to eq current_time

    expect(episode_record.body).to eq "にぱー"
    expect(episode_record.locale).to eq "ja"
    expect(episode_record.rating_state).to eq "good"
    expect(episode_record.episode_id).to eq episode.id
    expect(episode_record.record_id).to eq record.id
    expect(episode_record.work_id).to eq work.id

    expect(activity_group.itemable_type).to eq "EpisodeRecord"
    expect(activity_group.single).to eq true

    expect(activity.activity_group_id).to eq activity_group.id
    expect(activity.itemable).to eq episode_record

    expect(library_entry.work).to eq work
    expect(library_entry.watched_episode_ids).to eq [episode.id]
    expect(library_entry.next_episode).to eq next_episode
  end

  context "watched_at が指定されているとき" do
    let!(:watched_time) { Time.zone.parse("2021-01-01 12:00:00") }

    it "エピソードへの記録が作成できること" do
      form = Forms::EpisodeRecordForm.new(user: user, episode: episode)
      form.attributes = {
        comment: "にぱー",
        rating: "good",
        share_to_twitter: false,
        watched_at: watched_time
      }
      expect(form.valid?).to eq true

      Creators::EpisodeRecordCreator.new(user: user, form: form).call

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
      let(:episode) { create :episode, episode_record_bodies_count: 1 }
      let(:work) { episode.work }
      # 感想付きの記録が直前にある
      let(:episode_record) { create(:episode_record, user: user, episode: episode, body: "はうー") }
      let!(:activity_group) { create(:activity_group, user: user, itemable_type: "EpisodeRecord", single: true) }
      let!(:activity) { create(:activity, user: user, activity_group: activity_group, itemable: episode_record) }

      it "ActivityGroup が新たに作成されること" do
        expect(Record.count).to eq 1
        expect(EpisodeRecord.count).to eq 1
        expect(ActivityGroup.count).to eq 1
        expect(Activity.count).to eq 1
        expect(user.share_record_to_twitter?).to eq false

        # Creatorを呼ぶ
        form = Forms::EpisodeRecordForm.new(user: user, episode: episode)
        form.attributes = {
          comment: "にぱー", # 感想付きの記録を新たにする
          rating: "good",
          share_to_twitter: false
        }
        expect(form.valid?).to eq true

        Creators::EpisodeRecordCreator.new(user: user, form: form).call

        expect(ActivityGroup.count).to eq 2 # ActivityGroup が新たに作成されるはず
        expect(Activity.count).to eq 2

        episode_record = user.episode_records.last
        activity_group = user.activity_groups.last
        activity = user.activities.last

        expect(activity_group.itemable_type).to eq "EpisodeRecord"
        expect(activity_group.single).to eq true

        expect(activity.activity_group_id).to eq activity_group.id
        expect(activity.itemable).to eq episode_record
      end
    end

    context "直前の記録に感想が書かれていない & その後に新たに感想無しの記録をしたとき" do
      let(:user) { create :registered_user }
      let(:episode) { create :episode }
      let(:work) { episode.work }
      # 感想無しの記録が直前にある
      let(:episode_record) { create(:episode_record, user: user, episode: episode, body: "") }
      let!(:activity_group) { create(:activity_group, user: user, itemable_type: "EpisodeRecord", single: false) }
      let!(:activity) { create(:activity, user: user, activity_group: activity_group, itemable: episode_record) }

      it "ActivityGroup が新たに作成されないこと" do
        expect(Record.count).to eq 1
        expect(EpisodeRecord.count).to eq 1
        expect(ActivityGroup.count).to eq 1
        expect(Activity.count).to eq 1
        expect(user.share_record_to_twitter?).to eq false

        # Creatorを呼ぶ
        form = Forms::EpisodeRecordForm.new(user: user, episode: episode)
        form.attributes = {
          comment: "", # 感想無しの記録を新たにする
          rating: "good",
          share_to_twitter: false
        }
        expect(form.valid?).to eq true

        Creators::EpisodeRecordCreator.new(user: user, form: form).call

        expect(ActivityGroup.count).to eq 1 # ActivityGroup は新たに作成されないはず
        expect(Activity.count).to eq 2

        episode_record = user.episode_records.last
        activity_group = user.activity_groups.first
        activity = user.activities.last

        expect(activity_group.itemable_type).to eq "EpisodeRecord"
        expect(activity_group.single).to eq false

        expect(activity.itemable).to eq episode_record
        # もともとあった ActivityGroup に紐付くはず
        expect(activity.activity_group_id).to eq activity_group.id
      end
    end
  end
end
