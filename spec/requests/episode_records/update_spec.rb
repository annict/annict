# typed: false
# frozen_string_literal: true

RSpec.describe "PATCH /episode_records", type: :request do
  it "未ログインのとき、ログインページにリダイレクトすること" do
    patch episode_record_mutation_path

    expect(response).to have_http_status(:redirect)
    expect(response).to redirect_to(new_user_session_path)
  end

  it "ログイン済みのとき、エピソード記録を更新すること" do
    user = FactoryBot.create(:user)
    login_as(user, scope: :user)

    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:)
    episode_record = FactoryBot.create(:episode_record, user:, work:, episode:)

    params = {
      episode_record: {
        id: episode_record.id,
        comment: "面白かった！",
        rating: "great"
      }
    }

    patch episode_record_mutation_path, params: params

    expect(response).to have_http_status(:ok)

    episode_record.reload
    expect(episode_record.comment).to eq("面白かった！")
    expect(episode_record.rating_state).to eq("great")
  end

  it "ログイン済みで他のユーザーの記録のとき、403を返すこと" do
    user = FactoryBot.create(:user)
    other_user = FactoryBot.create(:user)
    login_as(user, scope: :user)

    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:)
    episode_record = FactoryBot.create(:episode_record, user: other_user, work:, episode:)

    params = {
      episode_record: {
        id: episode_record.id,
        comment: "面白かった！",
        rating: "great"
      }
    }

    patch episode_record_mutation_path, params: params

    expect(response).to have_http_status(:forbidden)

    episode_record.reload
    expect(episode_record.comment).not_to eq("面白かった！")
  end

  it "ログイン済みで無効なパラメータのとき、エラーを返すこと" do
    user = FactoryBot.create(:user)
    login_as(user, scope: :user)

    work = FactoryBot.create(:work)
    episode = FactoryBot.create(:episode, work:)
    episode_record = FactoryBot.create(:episode_record, user:, work:, episode:)

    params = {
      episode_record: {
        id: episode_record.id,
        comment: "",
        rating: "invalid_rating"
      }
    }

    patch episode_record_mutation_path, params: params

    expect(response).to have_http_status(:unprocessable_entity)
  end

  it "ログイン済みで記録が見つからないとき、404を返すこと" do
    user = FactoryBot.create(:user)
    login_as(user, scope: :user)

    params = {
      episode_record: {
        id: "invalid-id",
        comment: "面白かった！",
        rating: "great"
      }
    }

    patch episode_record_mutation_path, params: params

    expect(response).to have_http_status(:not_found)
  end
end
