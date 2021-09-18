# frozen_string_literal: true

describe Updaters::EpisodeRecordUpdater, type: :model do
  let(:user) { create :registered_user }
  let(:episode) { create :episode }
  let(:work) { episode.work }
  let!(:record) { create :record, :on_episode, user: user, work: work, episode: episode }
  let!(:episode_record) { record.episode_record }

  it "エピソードへの記録の更新ができること" do
    # 各レコードは1件のはず
    expect(Record.count).to eq 1
    expect(EpisodeRecord.count).to eq 1

    # Updaterを呼ぶ
    Updaters::EpisodeRecordUpdater.new(
      user: user,
      form: EpisodeRecordForm.new(
        user: user,
        body: record.body + "！！",
        episode: episode,
        rating: "good",
        record: record,
        share_to_twitter: false
      )
    ).call

    # Updaterを呼んでも各レコードは1件のまま
    expect(Record.count).to eq 1
    expect(EpisodeRecord.count).to eq 1

    record = user.records.first
    episode_record = record.episode_record

    expect(record.body).to eq "おもしろかった！！"
    expect(record.locale).to eq "ja"
    expect(record.rating).to eq "good"
    expect(record.episode_id).to eq episode.id
    expect(record.work_id).to eq work.id

    expect(episode_record).not_to be_nil
  end
end
