# typed: false
# frozen_string_literal: true

RSpec.describe "POST /api/internal/works/:work_id/commented_records", type: :request do
  it "未認証ユーザーの場合、リダイレクトされること" do
    work = FactoryBot.create(:work)

    post "/api/internal/works/#{work.id}/commented_records", params: {
      forms_work_record_form: {
        comment: "テストコメント",
        rating_overall: "good",
        watched_at: Time.zone.now
      }
    }

    expect(response.status).to eq(302)
  end

  it "存在しないworkの場合、404エラーが発生すること" do
    user = FactoryBot.create(:user, :with_profile)
    login_as(user, scope: :user)

    expect {
      post "/api/internal/works/99999/commented_records", params: {
        forms_work_record_form: {
          comment: "テストコメント",
          rating_overall: "good",
          watched_at: Time.zone.now
        }
      }
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除されたworkの場合、404エラーが発生すること" do
    user = FactoryBot.create(:user, :with_profile)
    work = FactoryBot.create(:work, deleted_at: Time.zone.now)
    login_as(user, scope: :user)

    expect {
      post "/api/internal/works/#{work.id}/commented_records", params: {
        forms_work_record_form: {
          comment: "テストコメント",
          rating_overall: "good",
          watched_at: Time.zone.now
        }
      }
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "有効なパラメータでワークレコードを作成できること" do
    user = FactoryBot.create(:user, :with_profile)
    work = FactoryBot.create(:work)
    login_as(user, scope: :user)

    expect {
      post "/api/internal/works/#{work.id}/commented_records", params: {
        forms_work_record_form: {
          comment: "素晴らしいアニメでした",
          rating_overall: "great",
          rating_animation: "good",
          rating_character: "average",
          rating_story: "good",
          rating_music: "great",
          watched_at: Time.zone.now
        }
      }
    }.to change(Record, :count).by(1)

    expect(response.status).to eq(201)
    expect(JSON.parse(response.body)).to eq({})

    record = Record.last
    expect(record.user).to eq(user)
    expect(record.work).to eq(work)
    work_record = record.work_record
    expect(work_record.body).to eq("素晴らしいアニメでした")
    expect(work_record.rating_overall_state).to eq("great")
    expect(work_record.rating_animation_state).to eq("good")
    expect(work_record.rating_character_state).to eq("average")
    expect(work_record.rating_story_state).to eq("good")
    expect(work_record.rating_music_state).to eq("great")
  end

  it "最小限のパラメータでワークレコードを作成できること" do
    user = FactoryBot.create(:user, :with_profile)
    work = FactoryBot.create(:work)
    login_as(user, scope: :user)

    expect {
      post "/api/internal/works/#{work.id}/commented_records", params: {
        forms_work_record_form: {
          comment: "短いコメント",
          watched_at: Time.zone.now
        }
      }
    }.to change(Record, :count).by(1)

    expect(response.status).to eq(201)
    expect(JSON.parse(response.body)).to eq({})

    record = Record.last
    expect(record.user).to eq(user)
    expect(record.work).to eq(work)
    work_record = record.work_record
    expect(work_record.body).to eq("短いコメント")
    expect(work_record.rating_overall_state).to be_nil
    expect(work_record.rating_animation_state).to be_nil
    expect(work_record.rating_character_state).to be_nil
    expect(work_record.rating_story_state).to be_nil
    expect(work_record.rating_music_state).to be_nil
  end

  it "空のコメントでワークレコードを作成できること" do
    user = FactoryBot.create(:user, :with_profile)
    work = FactoryBot.create(:work)
    login_as(user, scope: :user)

    expect {
      post "/api/internal/works/#{work.id}/commented_records", params: {
        forms_work_record_form: {
          comment: "",
          rating_overall: "good",
          watched_at: Time.zone.now
        }
      }
    }.to change(Record, :count).by(1)

    expect(response.status).to eq(201)
    record = Record.last
    expect(record.work_record.body).to eq("")
  end

  it "無効なパラメータの場合、422エラーを返すこと" do
    user = FactoryBot.create(:user, :with_profile)
    work = FactoryBot.create(:work)
    login_as(user, scope: :user)

    expect {
      post "/api/internal/works/#{work.id}/commented_records", params: {
        forms_work_record_form: {
          comment: "a" * 1_048_597, # 文字数制限を超える
          rating_overall: "good",
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
    work = FactoryBot.create(:work)
    login_as(user, scope: :user)

    expect {
      post "/api/internal/works/#{work.id}/commented_records", params: {
        forms_work_record_form: {
          comment: "テストコメント",
          rating_overall: "invalid_rating",
          watched_at: Time.zone.now
        }
      }
    }.not_to change(Record, :count)

    expect(response.status).to eq(422)
    expect(JSON.parse(response.body)).to be_an(Array)
  end

  it "評価値が大文字でも正常に処理されること" do
    user = FactoryBot.create(:user, :with_profile)
    work = FactoryBot.create(:work)
    login_as(user, scope: :user)

    expect {
      post "/api/internal/works/#{work.id}/commented_records", params: {
        forms_work_record_form: {
          comment: "テストコメント",
          rating_overall: "GREAT",
          rating_animation: "GOOD",
          watched_at: Time.zone.now
        }
      }
    }.to change(Record, :count).by(1)

    expect(response.status).to eq(201)
    record = Record.last
    work_record = record.work_record
    expect(work_record.rating_overall_state).to eq("great")
    expect(work_record.rating_animation_state).to eq("good")
  end
end
