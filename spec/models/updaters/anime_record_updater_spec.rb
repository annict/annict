# frozen_string_literal: true

describe Updaters::AnimeRecordUpdater, type: :model do
  let!(:user) { create :registered_user }
  let!(:anime) { create :anime }
  let!(:record) { create :record, user: user, anime: anime }
  let!(:anime_record) { create :anime_record, user: user, anime: anime, record: record }

  it "アニメへの記録の更新ができること" do
    # 各レコードは1件のはず
    expect(Record.count).to eq 1
    expect(WorkRecord.count).to eq 1

    # Updaterを呼ぶ
    Updaters::AnimeRecordUpdater.new(
      user: user,
      form: Forms::AnimeRecordForm.new(
        anime: anime,
        comment: anime_record.body + "！！",
        rating_animation: "great",
        rating_character: "great",
        rating_music: "great",
        rating_overall: "great",
        rating_story: "great",
        record: record,
        share_to_twitter: false
      )
    ).call

    # Updaterを呼んでも各レコードは1件のまま
    expect(Record.count).to eq 1
    expect(WorkRecord.count).to eq 1

    record = user.records.first
    anime_record = user.anime_records.first

    expect(record.work_id).to eq anime.id

    expect(anime_record.body).to eq "おもしろかった！！"
    expect(anime_record.locale).to eq "ja"
    expect(anime_record.rating_overall_state).to eq "great"
    expect(anime_record.rating_animation_state).to eq "great"
    expect(anime_record.rating_character_state).to eq "great"
    expect(anime_record.rating_music_state).to eq "great"
    expect(anime_record.rating_story_state).to eq "great"
    expect(anime_record.record_id).to eq record.id
    expect(anime_record.work_id).to eq anime.id
  end
end
