# typed: false
# frozen_string_literal: true

RSpec.describe "POST /api/internal/episodes/:episode_id/commented_records", type: :request do
  it "未認証ユーザーの場合、リダイレクトされること" do
    episode = FactoryBot.create(:episode)

    post "/api/internal/episodes/#{episode.id}/commented_records", params: {
      forms_episode_record_form: {
        comment: "テストコメント",
        rating: "good",
        watched_at: Time.zone.now
      }
    }

    expect(response.status).to eq(302)
  end

  it "存在しないエピソードの場合、404エラーが発生すること" do
    user = FactoryBot.create(:user, :with_profile)
    login_as(user, scope: :user)

    post "/api/internal/episodes/99999/commented_records", params: {
      forms_episode_record_form: {
        comment: "テストコメント",
        rating: "good",
        watched_at: Time.zone.now
      }
    }

    expect(response.status).to eq(404)
  end

  it "削除されたエピソードの場合、404エラーが発生すること" do
    user = FactoryBot.create(:user, :with_profile)
    episode = FactoryBot.create(:episode, deleted_at: Time.zone.now)
    login_as(user, scope: :user)

    post "/api/internal/episodes/#{episode.id}/commented_records", params: {
      forms_episode_record_form: {
        comment: "テストコメント",
        rating: "good",
        watched_at: Time.zone.now
      }
    }

    expect(response.status).to eq(404)
  end

  it "有効なパラメータでエピソードレコードを作成できること" do
    user = FactoryBot.create(:user, :with_profile)
    episode = FactoryBot.create(:episode)
    login_as(user, scope: :user)

    expect {
      post "/api/internal/episodes/#{episode.id}/commented_records", params: {
        forms_episode_record_form: {
          comment: "素晴らしいエピソードでした",
          rating: "great",
          watched_at: Time.zone.now
        }
      }
    }.to change(Record, :count).by(1)

    expect(response.status).to eq(201)
    expect(JSON.parse(response.body)).to eq({})

    record = Record.last
    expect(record.user).to eq(user)
    expect(record.episode).to eq(episode)
    episode_record = record.episode_record
    expect(episode_record.body).to eq("素晴らしいエピソードでした")
    expect(episode_record.rating_state).to eq("great")
  end

  it "最小限のパラメータでエピソードレコードを作成できること" do
    user = FactoryBot.create(:user, :with_profile)
    episode = FactoryBot.create(:episode)
    login_as(user, scope: :user)

    expect {
      post "/api/internal/episodes/#{episode.id}/commented_records", params: {
        forms_episode_record_form: {
          comment: "短いコメント",
          watched_at: Time.zone.now
        }
      }
    }.to change(Record, :count).by(1)

    expect(response.status).to eq(201)
    expect(JSON.parse(response.body)).to eq({})

    record = Record.last
    expect(record.user).to eq(user)
    expect(record.episode).to eq(episode)
    episode_record = record.episode_record
    expect(episode_record.body).to eq("短いコメント")
    expect(episode_record.rating_state).to be_nil
  end

  it "空のコメントでエピソードレコードを作成できること" do
    user = FactoryBot.create(:user, :with_profile)
    episode = FactoryBot.create(:episode)
    login_as(user, scope: :user)

    expect {
      post "/api/internal/episodes/#{episode.id}/commented_records", params: {
        forms_episode_record_form: {
          comment: "",
          rating: "good",
          watched_at: Time.zone.now
        }
      }
    }.to change(Record, :count).by(1)

    expect(response.status).to eq(201)
    record = Record.last
    expect(record.episode_record.body).to eq("")
  end

  it "無効なパラメータの場合、422エラーを返すこと" do
    user = FactoryBot.create(:user, :with_profile)
    episode = FactoryBot.create(:episode)
    login_as(user, scope: :user)

    expect {
      post "/api/internal/episodes/#{episode.id}/commented_records", params: {
        forms_episode_record_form: {
          comment: "a" * 1_048_597, # 文字数制限を超える
          rating: "good",
          watched_at: Time.zone.now
        }
      }
    }.not_to change(Record, :count)

    expect(response.status).to eq(422)
    expect(JSON.parse(response.body)).to be_an(Array)
    expect(JSON.parse(response.body)).to include(match(/感想.*1048596文字以内/))
  end

  it "無効な評価値の場合、422エラーを返すこと" do
    user = FactoryBot.create(:user, :with_profile)
    episode = FactoryBot.create(:episode)
    login_as(user, scope: :user)

    expect {
      post "/api/internal/episodes/#{episode.id}/commented_records", params: {
        forms_episode_record_form: {
          comment: "テストコメント",
          rating: "invalid_rating",
          watched_at: Time.zone.now
        }
      }
    }.not_to change(Record, :count)

    expect(response.status).to eq(422)
    expect(JSON.parse(response.body)).to be_an(Array)
  end

  it "評価値が大文字でも正常に処理されること" do
    user = FactoryBot.create(:user, :with_profile)
    episode = FactoryBot.create(:episode)
    login_as(user, scope: :user)

    expect {
      post "/api/internal/episodes/#{episode.id}/commented_records", params: {
        forms_episode_record_form: {
          comment: "テストコメント",
          rating: "GREAT",
          watched_at: Time.zone.now
        }
      }
    }.to change(Record, :count).by(1)

    expect(response.status).to eq(201)
    record = Record.last
    episode_record = record.episode_record
    expect(episode_record.rating_state).to eq("great")
  end
end
