# frozen_string_literal: true

describe Updaters::EpisodeRecordUpdater, type: :model do
  let(:user) { create :registered_user }
  let(:episode) { create :episode }
  let(:anime) { episode.anime }
  let!(:record) { create :record, user: user, anime: anime }
  let!(:episode_record) { create :episode_record, user: user, anime: anime, episode: episode, record: record }

  it "エピソードへの記録の更新ができること" do
    # 各レコードは1件のはず
    expect(Record.count).to eq 1
    expect(EpisodeRecord.count).to eq 1

    # Updaterを呼ぶ
    Updaters::EpisodeRecordUpdater.new(
      user: user,
      form: Forms::EpisodeRecordForm.new(
        comment: episode_record.body + "！！",
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
    episode_record = user.episode_records.first

    expect(record.work_id).to eq anime.id

    expect(episode_record.body).to eq "おもしろかった！！"
    expect(episode_record.locale).to eq "ja"
    expect(episode_record.rating_state).to eq "good"
    expect(episode_record.episode_id).to eq episode.id
    expect(episode_record.record_id).to eq record.id
    expect(episode_record.work_id).to eq anime.id
  end
end
