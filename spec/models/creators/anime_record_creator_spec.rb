# frozen_string_literal: true

describe Creators::AnimeRecordCreator, type: :model do
  let(:user) { create :registered_user }
  let(:anime) { create :anime }

  it "アニメへの記録ができること" do
    # Creatorを呼んでいないので、各レコードは0件のはず
    expect(Record.count).to eq 0
    expect(WorkRecord.count).to eq 0
    expect(ActivityGroup.count).to eq 0
    expect(Activity.count).to eq 0
    expect(user.share_record_to_twitter?).to eq false

    # Creatorを呼ぶ
    Creators::AnimeRecordCreator.new(
      user: user,
      form: Forms::AnimeRecordForm.new(
        anime: anime,
        comment: "すごく面白かった。",
        rating_animation: "great",
        rating_character: "great",
        rating_music: "great",
        rating_overall: "great",
        rating_story: "great",
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
    anime_record = user.anime_records.first
    activity_group = user.activity_groups.first
    activity = user.activities.first

    expect(record.work_id).to eq anime.id

    expect(anime_record.body).to eq "すごく面白かった。"
    expect(anime_record.locale).to eq "ja"
    expect(anime_record.rating_overall_state).to eq "great"
    expect(anime_record.rating_animation_state).to eq "great"
    expect(anime_record.rating_character_state).to eq "great"
    expect(anime_record.rating_music_state).to eq "great"
    expect(anime_record.rating_story_state).to eq "great"
    expect(anime_record.record_id).to eq record.id
    expect(anime_record.work_id).to eq anime.id

    expect(activity_group.itemable_type).to eq "AnimeRecord"
    expect(activity_group.single).to eq true

    expect(activity.itemable).to eq anime_record
    expect(activity.activity_group_id).to eq activity_group.id
  end

  describe "アクティビティの作成" do
    context "直前の記録に感想が書かれていて、その後に新たに感想付きの記録をしたとき" do
      let(:anime) { create :anime, work_records_with_body_count: 1 }
      # 感想付きの記録が直前にある
      let!(:anime_record) { create(:anime_record, user: user, anime: anime, body: "さいこー") }
      let!(:activity_group) { create(:activity_group, user: user, itemable_type: "WorkRecord", single: true) }
      let!(:activity) { create(:activity, user: user, itemable: anime_record, activity_group: activity_group) }

      it "ActivityGroup が新たに作成されること" do
        expect(ActivityGroup.count).to eq 1
        expect(Activity.count).to eq 1

        # Creatorを呼ぶ
        Creators::AnimeRecordCreator.new(
          user: user,
          form: Forms::AnimeRecordForm.new(
            anime: anime,
            comment: "すごく面白かった。", # 感想付きの記録を新たにする
            rating_animation: "great",
            rating_character: "great",
            rating_music: "great",
            rating_overall: "great",
            rating_story: "great",
            share_to_twitter: false
          )
        ).call

        expect(ActivityGroup.count).to eq 2 # ActivityGroup が新たに作成されるはず
        expect(Activity.count).to eq 2

        anime_record = user.anime_records.last
        activity_group = user.activity_groups.last
        activity = user.activities.last

        expect(activity_group.itemable_type).to eq "AnimeRecord"
        expect(activity_group.single).to eq true

        expect(activity.itemable).to eq anime_record
        expect(activity.activity_group_id).to eq activity_group.id
      end
    end

    context "直前の記録に感想が書かれていない & その後に新たに感想無しの記録をしたとき" do
      let(:anime) { create :anime }
      # 感想無しの記録が直前にある
      let!(:anime_record) { create(:anime_record, user: user, anime: anime, body: "") }
      let!(:activity_group) { create(:activity_group, user: user, itemable_type: "AnimeRecord", single: false) }
      let!(:activity) { create(:activity, user: user, itemable: anime_record, activity_group: activity_group) }

      it "ActivityGroup が新たに作成されないこと" do
        expect(ActivityGroup.count).to eq 1
        expect(Activity.count).to eq 1

        # Creatorを呼ぶ
        Creators::AnimeRecordCreator.new(
          user: user,
          form: Forms::AnimeRecordForm.new(
            anime: anime,
            comment: "", # 感想無しの記録を新たにする
            rating_animation: "great",
            rating_character: "great",
            rating_music: "great",
            rating_overall: "great",
            rating_story: "great",
            share_to_twitter: false
          )
        ).call

        expect(ActivityGroup.count).to eq 1 # ActivityGroup は新たに作成されないはず
        expect(Activity.count).to eq 2

        anime_record = user.anime_records.last
        activity_group = user.activity_groups.first
        activity = user.activities.last

        expect(activity_group.itemable_type).to eq "AnimeRecord"
        expect(activity_group.single).to eq false

        expect(activity.itemable).to eq anime_record
        # もともとあった ActivityGroup に紐付くはず
        expect(activity.activity_group_id).to eq activity_group.id
      end
    end
  end
end
