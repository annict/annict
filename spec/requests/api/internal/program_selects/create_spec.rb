# typed: false
# frozen_string_literal: true

RSpec.describe "POST /api/internal/works/:work_id/program_select", type: :request do
  it "未認証ユーザーの場合、401エラーを返すこと" do
    work = FactoryBot.create(:work)

    post "/api/internal/works/#{work.id}/program_select", params: {
      program_id: "0"
    }

    expect(response.status).to eq(401)
  end

  it "存在しないworkの場合、404エラーが発生すること" do
    user = FactoryBot.create(:user, :with_profile)
    login_as(user, scope: :user)

    expect {
      post "/api/internal/works/99999/program_select", params: {
        program_id: "0"
      }
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除されたworkの場合、404エラーが発生すること" do
    user = FactoryBot.create(:user, :with_profile)
    work = FactoryBot.create(:work, deleted_at: Time.zone.now)
    login_as(user, scope: :user)

    expect {
      post "/api/internal/works/#{work.id}/program_select", params: {
        program_id: "0"
      }
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "program_id 0の場合、プログラムを未選択にできること" do
    user = FactoryBot.create(:user, :with_profile)
    work = FactoryBot.create(:work)
    FactoryBot.create(:library_entry, user:, work:, program: nil)
    login_as(user, scope: :user)

    allow(user).to receive(:save_program_to_library_entry!)

    post "/api/internal/works/#{work.id}/program_select", params: {
      program_id: "0"
    }

    expect(response.status).to eq(204)
    expect(user).to have_received(:save_program_to_library_entry!).with(work, nil)
  end

  it "有効なprogram_idの場合、プログラムを選択できること" do
    user = FactoryBot.create(:user, :with_profile)
    work = FactoryBot.create(:work)
    program = FactoryBot.create(:program, work:)
    FactoryBot.create(:library_entry, user:, work:, program: nil)
    login_as(user, scope: :user)

    allow(user).to receive(:save_program_to_library_entry!)

    post "/api/internal/works/#{work.id}/program_select", params: {
      program_id: program.id
    }

    expect(response.status).to eq(204)
    expect(user).to have_received(:save_program_to_library_entry!).with(work, program)
  end

  it "存在しないprogram_idの場合、404エラーが発生すること" do
    user = FactoryBot.create(:user, :with_profile)
    work = FactoryBot.create(:work)
    login_as(user, scope: :user)

    expect {
      post "/api/internal/works/#{work.id}/program_select", params: {
        program_id: "99999"
      }
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除されたプログラムの場合、404エラーが発生すること" do
    user = FactoryBot.create(:user, :with_profile)
    work = FactoryBot.create(:work)
    program = FactoryBot.create(:program, work:, deleted_at: Time.zone.now)
    login_as(user, scope: :user)

    expect {
      post "/api/internal/works/#{work.id}/program_select", params: {
        program_id: program.id
      }
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "他のworkのプログラムを指定した場合、404エラーが発生すること" do
    user = FactoryBot.create(:user, :with_profile)
    work = FactoryBot.create(:work)
    other_work = FactoryBot.create(:work)
    other_program = FactoryBot.create(:program, work: other_work)
    login_as(user, scope: :user)

    expect {
      post "/api/internal/works/#{work.id}/program_select", params: {
        program_id: other_program.id
      }
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
