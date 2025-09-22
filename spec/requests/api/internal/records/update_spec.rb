# typed: false
# frozen_string_literal: true

RSpec.describe "PATCH /api/internal/@:username/records/:record_id", type: :request do
  it "未認証ユーザーの場合、リダイレクトされること" do
    user = FactoryBot.create(:user, :with_profile)
    record = FactoryBot.create(:record, :with_episode_record, user: user)

    patch "/api/internal/@#{user.username}/records/#{record.id}", params: {
      forms_episode_record_form: {
        comment: "更新されたコメント",
        rating: "good",
        watched_at: Time.zone.now
      }
    }

    expect(response.status).to eq(302)
  end

  it "他のユーザーのレコードを更新しようとした場合、404を返すこと" do
    user = FactoryBot.create(:user, :with_profile)
    other_user = FactoryBot.create(:user, :with_profile)
    record = FactoryBot.create(:record, :with_episode_record, user: other_user)

    login_as(user, scope: :user)

    patch "/api/internal/@#{other_user.username}/records/#{record.id}", params: {
      forms_episode_record_form: {
        comment: "更新されたコメント",
        rating: "good",
        watched_at: Time.zone.now
      }
    }

    expect(response.status).to eq(404)
  end

  it "存在しないユーザーのレコードを更新しようとした場合、404を返すこと" do
    user = FactoryBot.create(:user, :with_profile)
    record = FactoryBot.create(:record, :with_episode_record, user: user)

    login_as(user, scope: :user)

    patch "/api/internal/@nonexistent/records/#{record.id}", params: {
      forms_episode_record_form: {
        comment: "更新されたコメント",
        rating: "good",
        watched_at: Time.zone.now
      }
    }

    expect(response.status).to eq(404)
  end

  it "存在しないレコードを更新しようとした場合、404を返すこと" do
    user = FactoryBot.create(:user, :with_profile)

    login_as(user, scope: :user)

    patch "/api/internal/@#{user.username}/records/99999", params: {
      forms_episode_record_form: {
        comment: "更新されたコメント",
        rating: "good",
        watched_at: Time.zone.now
      }
    }

    expect(response.status).to eq(404)
  end

  it "エピソードレコードを正常に更新できること" do
    user = FactoryBot.create(:user, :with_profile)
    episode = FactoryBot.create(:episode)
    record = FactoryBot.create(:record, :with_episode_record, user: user, episode: episode)

    login_as(user, scope: :user)
    patch "/api/internal/@#{user.username}/records/#{record.id}", params: {
      forms_episode_record_form: {
        comment: "更新されたコメント",
        rating: "good",
        watched_at: Time.zone.now
      }
    }

    expect(response.status).to eq(200)
    expect(JSON.parse(response.body)).to eq({})

    record.reload
    episode_record = record.episode_record
    expect(episode_record.body).to eq("更新されたコメント")
    expect(episode_record.rating_state).to eq("good")
  end

  it "エピソードレコードで無効なパラメータの場合、422を返すこと" do
    user = FactoryBot.create(:user, :with_profile)
    episode = FactoryBot.create(:episode)
    record = FactoryBot.create(:record, :with_episode_record, user: user, episode: episode)

    login_as(user, scope: :user)
    patch "/api/internal/@#{user.username}/records/#{record.id}", params: {
      forms_episode_record_form: {
        comment: "a" * 1_048_597, # 文字数制限を超える
        rating: "invalid_rating",
        watched_at: Time.zone.now
      }
    }

    expect(response.status).to eq(422)
    expect(JSON.parse(response.body)).to be_an(Array)
    expect(JSON.parse(response.body)).to include(match(/感想.*1048596文字以内/))
  end

  it "ワークレコードを正常に更新できること" do
    user = FactoryBot.create(:user, :with_profile)
    work = FactoryBot.create(:work)
    record = FactoryBot.create(:record, :with_work_record, user: user, work: work)

    login_as(user, scope: :user)
    patch "/api/internal/@#{user.username}/records/#{record.id}", params: {
      forms_work_record_form: {
        comment: "更新されたワークコメント",
        rating_overall: "great",
        rating_animation: "good",
        rating_character: "average",
        rating_story: "good",
        rating_music: "great",
        watched_at: Time.zone.now
      }
    }

    expect(response.status).to eq(200)
    expect(JSON.parse(response.body)).to eq({})

    record.reload
    work_record = record.work_record
    expect(work_record.body).to eq("更新されたワークコメント")
    expect(work_record.rating_overall_state).to eq("great")
    expect(work_record.rating_animation_state).to eq("good")
    expect(work_record.rating_character_state).to eq("average")
    expect(work_record.rating_story_state).to eq("good")
    expect(work_record.rating_music_state).to eq("great")
  end

  it "ワークレコードで無効なパラメータの場合、422を返すこと" do
    user = FactoryBot.create(:user, :with_profile)
    work = FactoryBot.create(:work)
    record = FactoryBot.create(:record, :with_work_record, user: user, work: work)

    login_as(user, scope: :user)
    patch "/api/internal/@#{user.username}/records/#{record.id}", params: {
      forms_work_record_form: {
        comment: "a" * 1_048_597, # 文字数制限を超える
        rating_overall: "invalid_rating",
        watched_at: Time.zone.now
      }
    }

    expect(response.status).to eq(422)
    expect(JSON.parse(response.body)).to be_an(Array)
    expect(JSON.parse(response.body)).to include(match(/感想.*1048596文字以内/))
  end

  it "エピソードレコードのコメントを空文字で更新できること" do
    user = FactoryBot.create(:user, :with_profile)
    episode = FactoryBot.create(:episode)
    record = FactoryBot.create(:record, :with_episode_record, user: user, episode: episode)

    login_as(user, scope: :user)
    patch "/api/internal/@#{user.username}/records/#{record.id}", params: {
      forms_episode_record_form: {
        comment: "",
        rating: "good",
        watched_at: Time.zone.now
      }
    }

    expect(response.status).to eq(200)
    record.reload
    expect(record.episode_record.body).to eq("")
  end

  it "ワークレコードのコメントを空文字で更新できること" do
    user = FactoryBot.create(:user, :with_profile)
    work = FactoryBot.create(:work)
    record = FactoryBot.create(:record, :with_work_record, user: user, work: work)

    login_as(user, scope: :user)
    patch "/api/internal/@#{user.username}/records/#{record.id}", params: {
      forms_work_record_form: {
        comment: "",
        rating_overall: "good",
        watched_at: Time.zone.now
      }
    }

    expect(response.status).to eq(200)
    record.reload
    expect(record.work_record.body).to eq("")
  end

  it "エピソードレコードの評価を空で更新できること" do
    user = FactoryBot.create(:user, :with_profile)
    episode = FactoryBot.create(:episode)
    record = FactoryBot.create(:record, :with_episode_record, user: user, episode: episode)

    login_as(user, scope: :user)
    patch "/api/internal/@#{user.username}/records/#{record.id}", params: {
      forms_episode_record_form: {
        comment: "コメント",
        rating: nil,
        watched_at: Time.zone.now
      }
    }

    expect(response.status).to eq(200)
    record.reload
    expect(record.episode_record.rating_state).to be_nil
  end

  it "ワークレコードの評価を空で更新できること" do
    user = FactoryBot.create(:user, :with_profile)
    work = FactoryBot.create(:work)
    record = FactoryBot.create(:record, :with_work_record, user: user, work: work)

    login_as(user, scope: :user)
    patch "/api/internal/@#{user.username}/records/#{record.id}", params: {
      forms_work_record_form: {
        comment: "コメント",
        rating_overall: nil,
        rating_animation: nil,
        rating_character: nil,
        rating_story: nil,
        rating_music: nil,
        watched_at: Time.zone.now
      }
    }

    expect(response.status).to eq(200)
    record.reload
    work_record = record.work_record
    expect(work_record.rating_overall_state).to be_nil
    expect(work_record.rating_animation_state).to be_nil
    expect(work_record.rating_character_state).to be_nil
    expect(work_record.rating_story_state).to be_nil
    expect(work_record.rating_music_state).to be_nil
  end
end
