# typed: false
# frozen_string_literal: true

RSpec.describe "POST /api/internal/multiple_episode_records", type: :request do
  it "未認証ユーザーの場合、リダイレクトされること" do
    episode1 = FactoryBot.create(:episode)
    episode2 = FactoryBot.create(:episode)

    post "/api/internal/multiple_episode_records", params: {
      episode_ids: [episode1.id, episode2.id]
    }

    expect(response.status).to eq(302)
  end

  it "エピソードIDが空の配列の場合、複数のエピソードレコードが作成されないこと" do
    user = FactoryBot.create(:user, :with_profile)

    login_as(user, scope: :user)

    expect {
      post "/api/internal/multiple_episode_records", params: {
        episode_ids: []
      }
    }.not_to change(Record, :count)

    expect(response.status).to eq(201)
  end

  it "存在しないエピソードIDが含まれている場合、存在するエピソードのみレコードが作成されること" do
    user = FactoryBot.create(:user, :with_profile)
    episode1 = FactoryBot.create(:episode)
    episode2 = FactoryBot.create(:episode)
    nonexistent_id = "nonexistent"

    login_as(user, scope: :user)

    expect {
      post "/api/internal/multiple_episode_records", params: {
        episode_ids: [episode1.id, nonexistent_id, episode2.id]
      }
    }.to change(Record, :count).by(2)

    expect(response.status).to eq(201)

    # 作成されたレコードを確認
    records = Record.where(user: user)
    expect(records.count).to eq(2)
    expect(records.map(&:episode)).to contain_exactly(episode1, episode2)
  end

  it "削除されたエピソードが含まれている場合、削除されていないエピソードのみレコードが作成されること" do
    user = FactoryBot.create(:user, :with_profile)
    episode1 = FactoryBot.create(:episode)
    episode2 = FactoryBot.create(:episode, deleted_at: Time.zone.now)
    episode3 = FactoryBot.create(:episode)

    login_as(user, scope: :user)

    expect {
      post "/api/internal/multiple_episode_records", params: {
        episode_ids: [episode1.id, episode2.id, episode3.id]
      }
    }.to change(Record, :count).by(2)

    expect(response.status).to eq(201)

    # 削除されていないエピソードのレコードのみ作成されていることを確認
    records = Record.where(user: user)
    expect(records.map(&:episode)).to contain_exactly(episode1, episode3)
  end

  it "有効なパラメータで複数のエピソードレコードを作成できること" do
    user = FactoryBot.create(:user, :with_profile)
    work = FactoryBot.create(:work)
    episode1 = FactoryBot.create(:episode, work: work, sort_number: 1)
    episode2 = FactoryBot.create(:episode, work: work, sort_number: 2)
    episode3 = FactoryBot.create(:episode, work: work, sort_number: 3)

    login_as(user, scope: :user)

    expect {
      post "/api/internal/multiple_episode_records", params: {
        episode_ids: [episode2.id, episode1.id, episode3.id]
      }
    }.to change(Record, :count).by(3)

    expect(response.status).to eq(201)

    # レコードが正しく作成されていることを確認
    records = Record.where(user: user).includes(:episode_record)
    expect(records.count).to eq(3)

    # sort_number順に処理されていることを確認
    sorted_episodes = records.map(&:episode).sort_by(&:sort_number)
    expect(sorted_episodes).to eq([episode1, episode2, episode3])

    # 各レコードにepisode_recordが作成されていることを確認
    records.each do |record|
      expect(record.episode_record).to be_present
    end
  end

  it "同じエピソードIDが複数含まれている場合でも、重複なくレコードが作成されること" do
    user = FactoryBot.create(:user, :with_profile)
    episode1 = FactoryBot.create(:episode)
    episode2 = FactoryBot.create(:episode)

    login_as(user, scope: :user)

    expect {
      post "/api/internal/multiple_episode_records", params: {
        episode_ids: [episode1.id, episode2.id, episode1.id]
      }
    }.to change(Record, :count).by(2)

    expect(response.status).to eq(201)

    # 重複なくレコードが作成されていることを確認
    records = Record.where(user: user)
    expect(records.map(&:episode)).to contain_exactly(episode1, episode2)
  end

  it "フォームバリデーションエラーがある場合、トランザクションがロールバックされること" do
    user = FactoryBot.create(:user, :with_profile)
    episode1 = FactoryBot.create(:episode)
    episode2 = FactoryBot.create(:episode)

    # 2番目のエピソードでフォームバリデーションエラーが発生するようにモック
    form_mock = instance_double(Forms::EpisodeRecordForm)
    call_count = 0
    allow(Forms::EpisodeRecordForm).to receive(:new) do |args|
      call_count += 1
      if call_count == 2
        allow(form_mock).to receive(:invalid?).and_return(true)
        allow(form_mock).to receive(:errors).and_return(
          instance_double(ActiveModel::Errors, full_messages: ["バリデーションエラーが発生しました"])
        )
        form_mock
      else
        Forms::EpisodeRecordForm.new(**args)
      end
    end

    login_as(user, scope: :user)

    expect {
      post "/api/internal/multiple_episode_records", params: {
        episode_ids: [episode1.id, episode2.id]
      }
    }.not_to change(Record, :count)

    expect(response.status).to eq(422)

    response_body = JSON.parse(response.body)
    expect(response_body).to eq(["バリデーションエラーが発生しました"])
  end

  it "既に同じエピソードのレコードが存在する場合でも、新しいレコードを作成できること" do
    user = FactoryBot.create(:user, :with_profile)
    episode1 = FactoryBot.create(:episode)
    episode2 = FactoryBot.create(:episode)

    # 既存のレコードを作成
    existing_record = FactoryBot.create(:record, :with_episode_record, user: user, episode: episode1)

    login_as(user, scope: :user)

    expect {
      post "/api/internal/multiple_episode_records", params: {
        episode_ids: [episode1.id, episode2.id]
      }
    }.to change(Record, :count).by(2)

    expect(response.status).to eq(201)

    # 新しいレコードが作成されていることを確認
    new_records = Record.where(user: user).where.not(id: existing_record.id)
    expect(new_records.count).to eq(2)
    expect(new_records.map(&:episode)).to contain_exactly(episode1, episode2)
  end
end
