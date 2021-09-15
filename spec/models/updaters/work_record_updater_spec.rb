# frozen_string_literal: true

describe Updaters::WorkRecordUpdater, type: :model do
  let!(:user) { create :registered_user }
  let!(:work) { create :work }
  let!(:work_record) { create :work_record }
  let!(:record) { create :record, :for_work, user: user, work: work, recordable: work_record }

  it "アニメへの記録の更新ができること" do
    # 各レコードは1件のはず
    expect(Record.count).to eq 1
    expect(WorkRecord.count).to eq 1

    # Updaterを呼ぶ
    Updaters::WorkRecordUpdater.new(
      user: user,
      form: Forms::WorkRecordForm.new(
        user: user,
        work: work,
        body: record.body + "！！",
        animation_rating: "great",
        character_rating: "great",
        music_rating: "great",
        story_rating: "great",
        rating: "great",
        record: record,
        share_to_twitter: false
      )
    ).call

    # Updaterを呼んでも各レコードは1件のまま
    expect(Record.count).to eq 1
    expect(WorkRecord.count).to eq 1

    record = user.records.first
    work_record = record.work_record

    expect(record.work_id).to eq work.id
    expect(record.body).to eq "おもしろかった！！"
    expect(record.locale).to eq "ja"
    expect(record.rating).to eq "great"
    expect(record.animation_rating).to eq "great"
    expect(record.character_rating).to eq "great"
    expect(record.music_rating).to eq "great"
    expect(record.story_rating).to eq "great"
    expect(record.work_id).to eq work.id

    expect(work_record).to_not be_nil
  end
end
