#shellcheck shell=bash disable=SC2016

Describe 'Pre-test preparations'
    Context 'Redis FLUSHALL'
        It 'purges whole Redis database'
            When call tests/helpers/RESET.sh
            The output should include "OK"
            The status should eq 0
        End
    End
End

Describe 'User existence and modification'
    Context 'add user1 without admin existing'
        It 'should fail'
            When call tests/helpers/user_add.sh user1 pass1
            The output should include "HTTP/2 423" # Locked
            The output should include "ERR: No 'admin' user present, cannot continue."
        End
    End
    Context 'add admin'
        It 'should pass'
            When call tests/helpers/user_add.sh admin adminpass
            The output should include "HTTP/2 201" # Created
            The output should include "OK: User 'admin' created."
        End
    End
    Context 'add user1'
        It 'should pass'
            When call tests/helpers/user_add.sh user1 pass1
            The output should include "HTTP/2 201" # Created
            The output should include "OK: User 'user1' created."
        End
    End
    Context 'add user2'
        It 'should pass'
            When call tests/helpers/user_add.sh user2 pass2
            The output should include "HTTP/2 201" # Created
            The output should include "OK: User 'user2' created."
        End
    End
    Context 'admin purge user2 with bad admin password'
        It 'should fail with unauthenticated'
            When call tests/helpers/admin_user_purge.sh user2 adminBADpass
            The output should include "HTTP/2 403" # Forbidden
            The output should include "ERR: Illegal user."
        End
    End
    Context 'admin purge user2 with good admin password'
        It 'should pass'
            When call tests/helpers/admin_user_purge.sh user2 adminpass
            The output should include "HTTP/2 200" # OK
            The output should include "OK: User purged."
        End
    End
End

Describe 'Environment handling'
    Context 'admin add env1'
        It 'should pass'
            When call tests/helpers/admin_env_create.sh env1 adminpass
            The output should include "HTTP/2 200"
            The output should include "OK: Environment created."
        End
    End
    Context 'admin add env2'
        It 'should pass'
            When call tests/helpers/admin_env_create.sh env2 adminpass
            The output should include "HTTP/2 200"
            The output should include "OK: Environment created."
        End
    End
    Context 'admin add env3'
        It 'should pass'
            When call tests/helpers/admin_env_create.sh env3 adminpass
            The output should include "HTTP/2 200"
            The output should include "OK: Environment created."
        End
    End
    Context 'lock env1'
        It 'should pass'
            When call tests/helpers/env_lock.sh env1 user1 pass1 20310101
            The output should include 'OK'
            The status should eq 0
        End
    End
End

# Describe 'Adding environments and locking them'
# # add host1 to env1 --> fail, env locked
#     Context ''
#         It ''
#             When call tests/helpers/host_lock.sh host1 user1 pass1 20320202
#             The output should include 'OK'
#             The status should eq 0
#         End
#     End
# # add host2 to env2 --> ok
#     Context ''
#         It ''
#             When call tests/helpers/host_lock.sh host2 user2 pass2 20320202
#             The output should include 'OK'
#             The status should eq 0
#         End
#     End
# End

# Describe 'Releasing environment with admin'
# # admin maintenance env1 --> ok (maint and terminate works for admin from any status)
#     Context ''
#         It ''
#             When call tests/helpers/admin_env_unlock.sh env1 adminpass
#             The output should include 'OK'
#             The status should eq 0
#         End
#     End
# End
# End