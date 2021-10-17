# frozen_string_literal: true

describe Updaters::WorkRecordUpdater, type: :model do
  let!(:user) { create :registered_user }
  let!(:work) { create :work }
  let!(:record) { create :record, user: user, work: work }
  let!(:work_record) { create :work_record, user: user, work: work, record: record }

  it "アニメへの記録の更新ができること" do
    # 各レコードは1件のはず
    expect(Record.count).to eq 1
    expect(WorkRecord.count).to eq 1

    # Updaterを呼ぶ
    form = Forms::WorkRecordForm.new(user: user, record: record, work: work)
    form.attributes = {
      comment: work_record.body + "！！",
      rating_animation: "great",
      rating_character: "great",
      rating_music: "great",
      rating_overall: "great",
      rating_story: "great",
      share_to_twitter: false
    }
    expect(form.valid?).to eq true

    Updaters::WorkRecordUpdater.new(user: user, form: form).call

    # Updaterを呼んでも各レコードは1件のまま
    expect(Record.count).to eq 1
    expect(WorkRecord.count).to eq 1

    record = user.records.first
    work_record = user.work_records.first

    expect(record.work_id).to eq work.id

    expect(work_record.body).to eq "おもしろかった！！"
    expect(work_record.locale).to eq "ja"
    expect(work_record.rating_overall_state).to eq "great"
    expect(work_record.rating_animation_state).to eq "great"
    expect(work_record.rating_character_state).to eq "great"
    expect(work_record.rating_music_state).to eq "great"
    expect(work_record.rating_story_state).to eq "great"
    expect(work_record.record_id).to eq record.id
    expect(work_record.work_id).to eq work.id
  end
end
