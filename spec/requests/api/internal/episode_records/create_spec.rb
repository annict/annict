# typed: false
# frozen_string_literal: true

RSpec.describe "POST /api/internal/episode_records", type: :request do
  it "未認証ユーザーの場合、リダイレクトされること" do
    episode = FactoryBot.create(:episode)

    post "/api/internal/episode_records", params: {
      episode_id: episode.id
    }

    expect(response.status).to eq(302)
  end

  it "存在しないエピソードIDの場合、404を返すこと" do
    user = FactoryBot.create(:user, :with_profile)

    login_as(user, scope: :user)

    post "/api/internal/episode_records", params: {
      episode_id: "nonexistent"
    }

    expect(response.status).to eq(404)
  end

  it "削除されたエピソードの場合、404を返すこと" do
    user = FactoryBot.create(:user, :with_profile)
    episode = FactoryBot.create(:episode, deleted_at: Time.zone.now)

    login_as(user, scope: :user)

    post "/api/internal/episode_records", params: {
      episode_id: episode.id
    }

    expect(response.status).to eq(404)
  end

  it "有効なパラメータでエピソードレコードを作成できること" do
    user = FactoryBot.create(:user, :with_profile)
    episode = FactoryBot.create(:episode)

    login_as(user, scope: :user)
    post "/api/internal/episode_records", params: {
      episode_id: episode.id
    }

    expect(response.status).to eq(201)

    response_body = JSON.parse(response.body)
    expect(response_body).to have_key("record_id")
    expect(response_body["record_id"]).to be_present

    record = Record.find(response_body["record_id"])
    expect(record.user).to eq(user)
    expect(record.episode).to eq(episode)
    expect(record.episode_record).to be_present
  end

  it "フォームが無効な場合、400を返すこと" do
    user = FactoryBot.create(:user, :with_profile)
    episode = FactoryBot.create(:episode)

    # Forms::EpisodeRecordFormのバリデーションを確認するため、
    # 実際の無効なケースをテストする必要があります
    form = instance_double(Forms::EpisodeRecordForm)
    allow(form).to receive(:invalid?).and_return(true)
    allow(form).to receive(:errors).and_return(
      instance_double(ActiveModel::Errors, full_messages: ["エラーメッセージ"])
    )
    allow(Forms::EpisodeRecordForm).to receive(:new).and_return(form)

    login_as(user, scope: :user)
    post "/api/internal/episode_records", params: {
      episode_id: episode.id
    }

    expect(response.status).to eq(400)

    response_body = JSON.parse(response.body)
    expect(response_body).to have_key("message")
    expect(response_body["message"]).to eq("エラーメッセージ")
  end

  it "既に同じエピソードのレコードが存在する場合でも、新しいレコードを作成できること" do
    user = FactoryBot.create(:user, :with_profile)
    episode = FactoryBot.create(:episode)

    # 既存のレコードを作成
    existing_record = FactoryBot.create(:record, :with_episode_record, user: user, episode: episode)

    login_as(user, scope: :user)
    post "/api/internal/episode_records", params: {
      episode_id: episode.id
    }

    expect(response.status).to eq(201)

    response_body = JSON.parse(response.body)
    new_record = Record.find(response_body["record_id"])

    # 新しいレコードが既存のレコードと異なることを確認
    expect(new_record.id).not_to eq(existing_record.id)
    expect(new_record.user).to eq(user)
    expect(new_record.episode).to eq(episode)
  end
end
