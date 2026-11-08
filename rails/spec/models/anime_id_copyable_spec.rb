# typed: false
# frozen_string_literal: true

describe AnimeIdCopyable do
  # 作品・エピソードの写し元にするanimeの行を作り、そのidを返す。
  def create_anime_id
    ActiveRecord::Base.connection.select_value("INSERT INTO animes (status) VALUES ('published') RETURNING id")
  end

  def create_mapped_work_record
    work_record = FactoryBot.create(:work)
    work_record.update_column(:anime_id, create_anime_id)
    work_record
  end

  def create_mapped_episode_record(work_record)
    episode_record = FactoryBot.create(:episode, work: work_record)
    episode_record.update_column(:anime_id, create_anime_id)
    episode_record
  end

  # 作品を指す行の作り方。モデルごとに、作品を受け取って行を作る。
  record_builders = {
    Cast => ->(work_record) { FactoryBot.create(:cast, work: work_record) },
    Staff => ->(work_record) { FactoryBot.create(:staff, work: work_record) },
    Trailer => ->(work_record) { FactoryBot.create(:trailer, work: work_record) },
    WorkImage => lambda do |work_record|
      work_image_record = WorkImage.new(work: work_record, user: FactoryBot.create(:user), copyright: "©", image_data: "{}")
      work_image_record.save!(validate: false)
      work_image_record
    end,
    VodTitle => lambda do |work_record|
      VodTitle.create!(channel: FactoryBot.create(:channel), work: work_record, code: "code", name: "VODタイトル")
    end,
    Slot => lambda do |work_record|
      FactoryBot.create(:slot, work: work_record, episode: nil, program: FactoryBot.create(:program, work: work_record))
    end,
    Program => ->(work_record) { FactoryBot.create(:program, work: work_record) },
    SeriesWork => ->(work_record) { FactoryBot.create(:series_work, work: work_record) }
  }

  record_builders.each do |model_class, build_record|
    it "#{model_class.name}を作るとき、作品のanime_idを写すこと" do
      work_record = create_mapped_work_record

      related_record = build_record.call(work_record)

      expect(related_record.reload.anime_id).to eq(work_record.anime_id)
    end
  end

  it "作品を変えたとき、変えた先の作品のanime_idを写すこと" do
    work_record = create_mapped_work_record
    other_work_record = create_mapped_work_record
    cast_record = FactoryBot.create(:cast, work: work_record)

    cast_record.update!(work: other_work_record)

    expect(cast_record.reload.anime_id).to eq(other_work_record.anime_id)
  end

  it "作品のanime_idがまだ無いとき、空のまま保存すること" do
    work_record = FactoryBot.create(:work)

    cast_record = FactoryBot.create(:cast, work: work_record)

    expect(cast_record.reload.anime_id).to be_nil
  end

  it "外部キーの変わらない保存では写さないこと" do
    work_record = create_mapped_work_record
    cast_record = FactoryBot.create(:cast, work: work_record)
    # 同期バッチが直す前の食い違いに見立てて、作品と違う値にしておく。
    stale_anime_id = create_anime_id
    cast_record.update_column(:anime_id, stale_anime_id)

    cast_record.reload.update!(name: "#{cast_record.name}改")

    expect(cast_record.reload.anime_id).to eq(stale_anime_id)
  end

  it "放送枠にエピソードを紐づけたとき、エピソードのanime_idを写し、紐づけを外すと空にすること" do
    work_record = create_mapped_work_record
    episode_record = create_mapped_episode_record(work_record)
    slot_record = FactoryBot.create(:slot, work: work_record, episode: nil, program: FactoryBot.create(:program, work: work_record))
    expect(slot_record.reload.episode_anime_id).to be_nil

    slot_record.update!(episode: episode_record)
    expect(slot_record.reload.episode_anime_id).to eq(episode_record.anime_id)
    expect(slot_record.anime_id).to eq(work_record.anime_id)

    slot_record.update!(episode: nil)
    expect(slot_record.reload.episode_anime_id).to be_nil
    expect(slot_record.anime_id).to eq(work_record.anime_id)
  end

  it "Annict DBの保存 (検証を飛ばす保存) でも写すこと" do
    work_record = create_mapped_work_record
    other_work_record = create_mapped_work_record
    episode_record = create_mapped_episode_record(other_work_record)
    slot_record = FactoryBot.create(:slot, work: work_record, episode: nil, program: FactoryBot.create(:program, work: work_record))

    slot_record.reload
    slot_record.user = FactoryBot.create(:registered_user)
    slot_record.work = other_work_record
    slot_record.episode = episode_record
    slot_record.save_and_create_activity!

    expect(slot_record.reload.anime_id).to eq(other_work_record.anime_id)
    expect(slot_record.episode_anime_id).to eq(episode_record.anime_id)
  end
end
