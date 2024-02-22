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
            The output should include '"success": false' # Locked
            The output should include "ERR: No 'admin' user present, cannot continue."
        End
    End
    Context 'add admin'
        It 'should pass'
            When call tests/helpers/user_add.sh admin adminpass
            The output should include '"success": true' # Created
            The output should include "OK: User 'admin' created."
        End
    End
    Context 'add user1'
        It 'should pass'
            When call tests/helpers/user_add.sh user1 pass1
            The output should include '"success": true' # Created
            The output should include "OK: User 'user1' created."
        End
    End
    Context 'add user1 again'
        It 'should fail, existing user'
            When call tests/helpers/user_add.sh user1 pass1
            The output should include '"success": false' # Forbidden
            The output should include "ERR: User already exists."
        End
    End
    Context 'add user2'
        It 'should pass'
            When call tests/helpers/user_add.sh user2 pass2
            The output should include '"success": true' # Created
            The output should include "OK: User 'user2' created."
        End
    End
    Context 'add user3'
        It 'should pass'
            When call tests/helpers/user_add.sh user3 pass3
            The output should include '"success": true' # Created
            The output should include "OK: User 'user3' created."
        End
    End
    Context 'admin purge user3 with bad admin password'
        It 'should fail with unauthenticated'
            When call tests/helpers/admin_user_purge.sh user3 adminBADpass
            The output should include '"success": false' # Forbidden
            The output should include "ERR: Illegal user."
        End
    End
    Context 'admin purge user3 with good admin password'
        It 'should pass'
            When call tests/helpers/admin_user_purge.sh user3 adminpass
            The output should include '"success": true' # OK
            The output should include "OK: User purged."
        End
    End
End

Describe 'Environment and host handling'
    Context 'admin create env1'
        It 'should pass'
            When call tests/helpers/admin_env_create.sh env1 adminpass
            The output should include '"success": true'
            The output should include "OK: Environment created."
        End
    End
    Context 'admin create env2'
        It 'should pass'
            When call tests/helpers/admin_env_create.sh env2 adminpass
            The output should include '"success": true'
            The output should include "OK: Environment created."
        End
    End
    Context 'admin create env3'
        It 'should pass'
            When call tests/helpers/admin_env_create.sh env3 adminpass
            The output should include '"success": true'
            The output should include "OK: Environment created."
        End
    End
    Context 'admin create env4'
        It 'should pass'
            When call tests/helpers/admin_env_create.sh env4 adminpass
            The output should include '"success": true'
            The output should include "OK: Environment created."
        End
    End
    Context 'admin create env5'
        It 'should pass'
            When call tests/helpers/admin_env_create.sh env5 adminpass
            The output should include '"success": true'
            The output should include "OK: Environment created."
        End
    End
    Context 'lock env1'
        It 'should pass'
            When call tests/helpers/env_lock.sh env1 user1 pass1 20310101
            The output should include '"success": true'
            The output should include "OK: Environment locked successfully."
        End
    End
    Context 'lock env1-host1'
        It 'should fail, env1 locked'
            When call tests/helpers/host_lock.sh env1-host1 user1 pass1 20310101
            The output should include '"success": false'
            The output should include "ERR: Parent environment is locked, cannot lock host."
        End
    End
    Context 'lock env2-host2'
        It 'should fail, deleted user'
            When call tests/helpers/host_lock.sh env2-host2 user3 pass3 20320202
            The output should include '"success": false'
            The output should include "ERR: Illegal user."
        End
    End
    Context 'lock env2-host2'
        It 'should fail, bad user password'
            When call tests/helpers/host_lock.sh env2-host2 user1 BADPASS 20320202
            The output should include '"success": false'
            The output should include "ERR: Illegal user."
        End
    End
    Context 'lock env2-host2'
        It 'should pass'
            When call tests/helpers/host_lock.sh env2-host2 user1 pass1 20320202
            The output should include '"success": true'
            The output should include "OK: Host has been locked succesfully."
        End
    End
    Context 'lock env2-host2'
        It 'should fail, locked by user1'
            When call tests/helpers/host_lock.sh env2-host2 user2 pass2 20320202
            The output should include '"success": false'
            The output should include "ERR:"
        End
    End
    Context 'lock env2-host3'
        It 'should pass'
            When call tests/helpers/host_lock.sh env2-host3 user1 pass1 20320202
            The output should include '"success": true'
            The output should include "OK: Host has been locked succesfully."
        End
    End
    Context 'lock env5-host1'
        It 'should pass'
            When call tests/helpers/host_lock.sh env5-host1-a user1 pass1 20320202
            The output should include '"success": true'
            The output should include "OK: Host has been locked succesfully."
        End
    End
    Context 'lock env6-host4'
        It 'should fail, no such env'
            When call tests/helpers/host_lock.sh env6-host4 user1 pass1 20320202
            The output should include '"success": false'
            The output should include "ERR: Parent env not defined, admin can add it."
        End
    End
    Context 'admin release env2'
        It 'should pass'
            When call tests/helpers/admin_env_unlock.sh env2 adminpass
            The output should include '"success": true'
            The output should include "OK: Environment unlocked."
        End
    End
    Context 'admin terminate env3'
        It 'should pass'
            When call tests/helpers/admin_env_terminate.sh env3 adminpass
            The output should include '"success": true'
            The output should include "OK: Environment terminated."
        End
    End
    Context 'admin maintenance env4'
        It 'should pass'
            When call tests/helpers/admin_env_maintenance.sh env4 adminpass
            The output should include '"success": true'
            The output should include "OK: Environment is in maintenance mode now."
        End
    End
End
